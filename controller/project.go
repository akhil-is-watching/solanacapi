package controller

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/PlenaFinance/solanacapi/helpers"
	"github.com/PlenaFinance/solanacapi/types"
	"github.com/gofiber/fiber/v2"
)

var (
	passingRe   = regexp.MustCompile(`(\d+) passing \((\d+ms)\)`)
	failingRe   = regexp.MustCompile(`(\d+) failing`)
	testSuiteRe = regexp.MustCompile(`^\s{2}(\w+)$`) // e.g., "  mata"
	// Updated to handle both ✔ and numbered failed tests
	testCaseRe = regexp.MustCompile(`^\s{4}(✔|\d+\))\s+(.+?)(?:\s+\((\d+ms)\))?$`)
)

func CompileProject(c *fiber.Ctx) error {
	var req types.CompileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	dir := "./workspace"
	programDir := dir + "/programs/" + req.ProjectName
	testDir := dir + "/tests/"

	// Ensure the program directory exists
	if err := os.MkdirAll(programDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create program directory: " + err.Error(),
		})
	}

	// Write each program file to the appropriate location
	for _, file := range req.ProgramFiles {
		filePath := programDir + "/" + file[0]
		if err := helpers.CreateFile(file[1], filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to write program file: " + err.Error(),
			})
		}
	}

	fmt.Println("Program files written")

	// Write each program config file to the appropriate location
	for _, file := range req.TestFiles {
		filePath := testDir + file[0]
		fmt.Println("Writing test file to", file[1])
		if err := helpers.CreateFile(file[1], filePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to write program config file: " + err.Error(),
			})
		}
	}

	// Write each test file to the appropriate location

	command := "anchor"
	args := []string{"build", "-p", req.ProjectName}

	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	// Capture stdout and stderr
	buildOutput, err := cmd.CombinedOutput()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to compile project: " + err.Error(),
			"output": string(buildOutput),
		})
	}

	command = "anchor"
	args = []string{"test", "-p", req.ProjectName}

	cmd = exec.Command(command, args...)
	cmd.Dir = dir

	// Capture stdout and stderr
	testOutput, err := cmd.CombinedOutput()
	testResults := parseTestLog(string(testOutput))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":       "Failed to test project: " + err.Error(),
			"testResults": testResults,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Project compiled successfully",
		"testResults": testResults,
	})
}

func parseTestLog(log string) map[string]interface{} {
	lines := strings.Split(log, "\n")

	testResults := map[string]interface{}{
		"summary": map[string]interface{}{
			"total":    0,
			"passed":   0,
			"failed":   0,
			"duration": "",
		},
		"tests":         []map[string]interface{}{},
		"executionLogs": lines,
		"rawOutput":     log,
	}

	// Parse summary
	for _, line := range lines {
		if matches := passingRe.FindStringSubmatch(line); matches != nil {
			summary := testResults["summary"].(map[string]interface{})
			summary["passed"] = atoi(matches[1])
			summary["duration"] = matches[2] + "ms"
			summary["total"] = summary["total"].(int) + atoi(matches[1])
		}
		if matches := failingRe.FindStringSubmatch(line); matches != nil {
			summary := testResults["summary"].(map[string]interface{})
			summary["failed"] = atoi(matches[1])
			summary["total"] = summary["total"].(int) + atoi(matches[1])
		}
	}

	// Parse tests
	tests := []map[string]interface{}{}
	var currentSuite string
	var currentTest map[string]interface{}
	var errorLines []string
	collectingError := false

	for i, line := range lines {
		// Test suite (2 spaces)
		if matches := testSuiteRe.FindStringSubmatch(line); matches != nil {
			currentSuite = matches[1]
			continue
		}

		// Test case (4 spaces, ✔ or number)
		if matches := testCaseRe.FindStringSubmatch(line); matches != nil {
			// Save previous test if exists
			if currentTest != nil {
				currentTest["error"] = strings.Join(errorLines, "\n")
				tests = append(tests, currentTest)
			}

			isPassed := matches[1] == "✔"
			duration := ""
			if len(matches) > 3 && matches[3] != "" {
				duration = matches[3]
			}

			currentTest = map[string]interface{}{
				"suite":    currentSuite,
				"name":     strings.TrimSpace(matches[2]),
				"passed":   isPassed,
				"error":    "",
				"duration": duration,
			}
			errorLines = []string{}
			collectingError = !isPassed // Only collect errors for failed tests
			continue
		}

		// Collect error details for failed tests
		if collectingError && currentTest != nil && strings.TrimSpace(line) != "" {
			// Stop collecting at summary or empty line after error
			if passingRe.MatchString(line) || failingRe.MatchString(line) {
				collectingError = false
				continue
			}
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" && strings.TrimSpace(line) == "" {
				collectingError = false
				continue
			}
			errorLines = append(errorLines, line)
		}
	}

	// Add the last test
	if currentTest != nil {
		currentTest["error"] = strings.Join(errorLines, "\n")
		tests = append(tests, currentTest)
	}

	testResults["tests"] = tests
	return testResults
}

// Helper function to convert string to int
func atoi(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
