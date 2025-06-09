package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"regexp"
	"strings"

	cdp "github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

func WriteUniqueGamesToFile(filename string, games []string) error {
	existing := make(map[string]struct{})
	if file, err := os.Open(filename); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				existing[line] = struct{}{}
			}
		}
		file.Close()
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, game := range games {
		game = strings.TrimSpace(game)
		if game == "" {
			continue
		}
		if _, found := existing[game]; !found {
			if _, err := file.WriteString(game + "\n"); err != nil {
				return err
			}
			existing[game] = struct{}{}
		}
	}
	return nil
}

func ExtractGamesBlock(codeText string) []string {
	lines := strings.Split(codeText, "\n")
	var games []string
	recording := false

	// Regex to match marker lines with optional leading numbers and spaces
	markerStart := regexp.MustCompile(`^\d*\s*playnitegames\s*$`)
	markerEnd := regexp.MustCompile(`^\d*\s*playnitegamesend\s*$`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if markerStart.MatchString(trimmed) {
			recording = true
			continue
		}
		if markerEnd.MatchString(trimmed) {
			break
		}
		if recording && trimmed != "" {
			games = append(games, trimmed)
		}
	}
	return games
}

func CleanGameNames(lines []string) []string {
	var cleaned []string
	re := regexp.MustCompile(`^\d+\s*`)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		cleanedLine := re.ReplaceAllString(trimmed, "")
		if cleanedLine != "" &&
			cleanedLine != "playnitegames" &&
			cleanedLine != "playnitegamesend" {
			cleaned = append(cleaned, cleanedLine)
		}
	}
	return cleaned
}

func ClickIfPresent(ctx context.Context, selector string) error {
	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Nodes(selector, &nodes, chromedp.ByQuery),
	)
	if err != nil {
		return err
	}
	if len(nodes) > 0 {

		return chromedp.Run(ctx,
			chromedp.Click(selector, chromedp.ByQuery),
		)
	}

	return nil
}

func main() {
	prompt := "Please extract the names of the games from the image(s) I provided. Output the result as a plain, deduplicated list that starts with the word 'playnitegames' (on its own line) also end it with 'playnitegamesend' (on its own line too), followed by one game name per line. Do not include any extra text or commentary—just the list. If a game name appears more than once, include it only once. Focus on upcoming or recently released games."
	imagePath := "C:/Users/Damasco/Pictures/photos/pic.jpg"
	chromeProfilePath := filepath.Join(
		os.Getenv("LOCALAPPDATA"),
		"Google",
		"Chrome",
		"User Data",
		"Default",
	)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(chromeProfilePath),
		chromedp.WindowSize(1280, 800),
		chromedp.Flag("headless", false),
		chromedp.Flag("start-minimized", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("profile-directory", "Default"),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
		chromedp.Flag("useAutomationExtension", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	err := chromedp.Run(ctx,

		chromedp.Navigate(`https://www.blackbox.ai`),
		chromedp.Sleep(4*time.Second),
		// chromedp.WaitVisible(`body`, chromedp.ByQuery),

		// chromedp.Click(`//button[.//span[text()='Upload']]`, chromedp.NodeVisible),
		chromedp.Click(`/html/body/div[2]/main/div/div[2]/div[2]/div/div/div/div/div/div/div/div[1]/div/div/div/div/div[1]/div/div[3]/div[1]/div/form/div[4]/div[2]/div[1]/button[5]`, chromedp.NodeVisible),
		chromedp.Sleep(2*time.Second),
		// chromedp.WaitVisible(`#file-input-RIb8Cof`, chromedp.ByID),

		chromedp.SetUploadFiles(`input[type="file"]`, []string{imagePath}, chromedp.ByQuery),

		chromedp.SendKeys(`#chat-input-box`, prompt, chromedp.ByID),
		chromedp.Sleep(5*time.Second),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
		chromedp.Click(`button[type="submit"]`, chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),

		// chromedp.Click(`//button[@id='prompt-form-send-button']`, chromedp.BySearch),
		// chromedp.SetValue(`/html/body/div[2]/main/div/div[2]/div[2]/div/div/div/div/div/div/div/div[1]/div/div/div/div/div[2]/form/div/textarea`, ""),
		// chromedp.SendKeys(`/html/body/div[2]/main/div/div[2]/div[2]/div/div/div/div/div/div/div/div[1]/div/div/div/div/div[2]/form/div/textarea`, prompt, chromedp.ByID),

	)
	if err != nil {
		log.Fatal(err)
	}

	selector := `button.inline-flex.items-center.justify-center.rounded-md.text-sm.font-medium.ring-offset-background.transition-colors.focus-visible\:outline-none.focus-visible\:ring-2.focus-visible\:ring-ring.focus-visible\:ring-offset-2.disabled\:pointer-events-none.disabled\:opacity-50.border.border-input.hover\:bg-accent.hover\:text-accent-foreground.h-8.px-4.py-2.w-full.shadow-none`

	err = ClickIfPresent(ctx, selector)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(12 * time.Second)
	var codeText string
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`div[style*="font-family: Consolas"] > code`, chromedp.ByQuery),
		chromedp.Text(`div[style*="font-family: Consolas"] > code`, &codeText, chromedp.NodeVisible, chromedp.ByQuery),
	)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println("Extracted code block:\n", codeText)

	gamesBlock := ExtractGamesBlock(codeText)
	cleanedGames := CleanGameNames(gamesBlock)

	fmt.Println("Games found:")
	for _, game := range cleanedGames {
		fmt.Println(game)
	}

	txtFile := "games.txt"
	err = WriteUniqueGamesToFile(txtFile, cleanedGames)
	if err != nil {
		log.Fatal("Failed to write games:", err)
	}
	fmt.Println("Games written to", txtFile)

	fmt.Println("Image upload complete!")
	time.Sleep(400 * time.Second)
}
