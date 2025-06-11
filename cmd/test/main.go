package main

import "github.com/fatih/color"

func main() {
	color.Red("We have red")

	c := color.New(color.FgCyan).Add(color.Underline)
	c.Println("Prints cyan text with an underline.")

	d := color.New(color.FgCyan, color.Bold)
	d.Printf("This prints bold cyan %s\n", "too!.")

	red := color.New(color.FgRed)
	boldRed := red.Add(color.Bold)
	boldRed.Println("This will print text in bold red.")

	whiteBackground := red.Add(color.BgHiBlue)
	whiteBackground.Println("Red text with white background.")

	f := color.New(color.FgRed).PrintfFunc()
	f("Warning\n")
}
