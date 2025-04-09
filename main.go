package main

import (
	"github.com/PlenaFinance/solanacapi/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	routes.SetupRoutes(app)
	app.Listen(":3000")
}

// func main() {
// 	// Define the directory where the command should be executed
// 	directory := "./sample"

// 	// Define the command and its arguments
// 	command := "anchor"
// 	args := []string{"test"}

// 	// Create a new command
// 	cmd := exec.Command(command, args...)

// 	// Set the working directory for the command
// 	cmd.Dir = directory

// 	// Set the command's stdout and stderr to the current process's
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	// Execute the command
// 	fmt.Printf("Executing '%s %s' in directory: %s\n", command, filepath.Join(args...), directory)
// 	err := cmd.Run()
// 	if err != nil {
// 		fmt.Printf("Error executing command: %v\n", err)
// 		os.Exit(1)
// 	}
// }
