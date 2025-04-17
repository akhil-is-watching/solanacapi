package controller

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PlenaFinance/solanacapi/types"
	"github.com/gagliardetto/solana-go"
	"github.com/gofiber/fiber/v2"
)

func Test(ctx *fiber.Ctx) error {
	// Use current directory instead of temp dir
	currentDir, err := os.Getwd()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to get current directory: %v", err),
		})
	}
	anchorWorkspace := filepath.Join(currentDir, "anchor-workspace")
	fmt.Println(anchorWorkspace)

	// Step 1: Parse the request body
	var req types.TestRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Step 2: Create project directory and write the necessary files to anchor-workspace
	projectDir := filepath.Join(anchorWorkspace, "programs", req.ProjectName)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create project directory: %v", err),
		})
	}
	for _, file := range req.ProgramFiles {
		// Create any directories that are necessary
		os.MkdirAll(filepath.Join(projectDir, filepath.Dir(file[0])), 0755)
		os.WriteFile(filepath.Join(projectDir, file[0]), []byte(file[1]), 0644)
	}

	// Write test files to the tests directory
	testsDir := filepath.Join(anchorWorkspace, "tests")
	if err := os.MkdirAll(testsDir, 0755); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create tests directory: %v", err),
		})
	}
	for _, file := range req.TestFiles {
		// Create any directories that are necessary
		os.MkdirAll(filepath.Join(testsDir, filepath.Dir(file[0])), 0755)
		os.WriteFile(filepath.Join(testsDir, file[0]), []byte(file[1]), 0644)
	}

	// Step 3: Create a new solana keypair and write it to anchor-workspace
	keypair, err := solana.NewRandomPrivateKey()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create the target directory if it doesn't exist
	targetDir := filepath.Join(anchorWorkspace, "target/deploy")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create directory: %v", err),
		})
	}

	// Write the private key in the required format
	keypairPath := filepath.Join(targetDir, fmt.Sprintf("%s-keypair.json", req.ProjectName))
	keypairBytes := []byte(keypair)

	// Format the keypair bytes as a JSON array with commas
	var formattedBytes strings.Builder
	formattedBytes.WriteString("[")
	for i, b := range keypairBytes {
		if i > 0 {
			formattedBytes.WriteString(", ")
		}
		formattedBytes.WriteString(fmt.Sprintf("%d", b))
	}
	formattedBytes.WriteString("]")

	if err := os.WriteFile(keypairPath, []byte(formattedBytes.String()), 0600); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to write keypair: %v", err),
		})
	}

	// Store the public key for later use
	publicKey := keypair.PublicKey().String()

	// Step 4: Replace the program id where necessary with the new keypair
	// 4.1: Update Anchor.toml
	anchorTomlPath := filepath.Join(anchorWorkspace, "Anchor.toml")

	// Read the Anchor.toml file
	anchorTomlContent, err := os.ReadFile(anchorTomlPath)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read Anchor.toml: %v", err),
		})
	}

	// Check if an entry with the same project name already exists
	content := string(anchorTomlContent)
	projectRegex := regexp.MustCompile(fmt.Sprintf(`%s = "([^"]+)"`, regexp.QuoteMeta(req.ProjectName)))
	matches := projectRegex.FindStringSubmatch(content)

	var existingProgramId string
	if len(matches) > 1 {
		// Entry exists, get the existing program ID
		existingProgramId = matches[1]
		// We'll use this existing program ID instead of generating a new one
		publicKey = existingProgramId
	} else {
		// No existing entry, add new entry to [programs.localnet] section
		re := regexp.MustCompile(`(\[programs\.localnet\](?:\s*[a-zA-Z0-9_-]+ = "[^"]+"\s*)*)`)
		newEntry := fmt.Sprintf("\n%s = \"%s\"", req.ProjectName, publicKey)
		updatedContent := re.ReplaceAllString(content, "${1}"+newEntry)

		if err := os.WriteFile(anchorTomlPath, []byte(updatedContent), 0644); err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update Anchor.toml: %v", err),
			})
		}
	}

	// 4.2: Update the program ID in the Rust source file
	libRsPath := filepath.Join(anchorWorkspace, "programs", req.ProjectName, "src", "lib.rs")

	// Read the lib.rs file
	libRsContent, err := os.ReadFile(libRsPath)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read lib.rs: %v", err),
		})
	}

	// Replace the program ID in the declare_id! macro using regex
	oldContent := string(libRsContent)
	re := regexp.MustCompile(`declare_id!\s*\(\s*"[^"]+"\s*\)`)
	updatedLibRs := re.ReplaceAllString(oldContent, fmt.Sprintf(`declare_id!("%s")`, publicKey))

	// Write the updated content back to lib.rs
	if err := os.WriteFile(libRsPath, []byte(updatedLibRs), 0644); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update lib.rs: %v", err),
		})
	}

	// Step 5: Run the anchor build command
	command := "anchor"
	args := []string{"build", "-p", req.ProjectName}

	cmd := exec.Command(command, args...)
	cmd.Dir = anchorWorkspace

	buildOutput, err := cmd.CombinedOutput()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to compile project: " + err.Error(),
			"output": string(buildOutput),
		})
	}

	// Step 6: Run the anchor test command
	command = "anchor"
	args = []string{"test", "-p", req.ProjectName}

	cmd = exec.Command(command, args...)
	cmd.Dir = anchorWorkspace

	testOutput, err := cmd.CombinedOutput()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to test project: " + err.Error(),
			"output": string(testOutput),
		})
	}

	// Step 7: Return the test results
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"output": string(testOutput),
	})
}
