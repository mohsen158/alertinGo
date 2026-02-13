package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	deployDir := os.Getenv("DEPLOY_DIR")
	if deployDir == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}
		deployDir = dir
	}

	branch := os.Getenv("DEPLOY_BRANCH")
	if branch == "" {
		branch = "main"
	}

	intervalSec := 30
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 5 {
			log.Fatalf("POLL_INTERVAL must be a number >= 5, got: %s", v)
		}
		intervalSec = n
	}

	scriptPath := filepath.Join(deployDir, "scripts", "deploy.sh")

	log.Printf("Deploy poller started: dir=%s branch=%s interval=%ds\n", deployDir, branch, intervalSec)

	for {
		if hasNewCommits(deployDir, branch) {
			log.Println("New commits detected, starting deploy...")
			runDeploy(scriptPath, deployDir)
		}
		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}

func hasNewCommits(dir, branch string) bool {
	// Fetch latest from remote
	fetch := exec.Command("git", "fetch", "origin", branch)
	fetch.Dir = dir
	if out, err := fetch.CombinedOutput(); err != nil {
		log.Printf("git fetch failed: %v\n%s", err, out)
		return false
	}

	// Compare local HEAD with remote
	local := exec.Command("git", "rev-parse", "HEAD")
	local.Dir = dir
	localOut, err := local.Output()
	if err != nil {
		log.Printf("git rev-parse HEAD failed: %v", err)
		return false
	}

	remote := exec.Command("git", "rev-parse", "origin/"+branch)
	remote.Dir = dir
	remoteOut, err := remote.Output()
	if err != nil {
		log.Printf("git rev-parse origin/%s failed: %v", branch, err)
		return false
	}

	localHash := strings.TrimSpace(string(localOut))
	remoteHash := strings.TrimSpace(string(remoteOut))

	if localHash != remoteHash {
		log.Printf("Local: %s Remote: %s\n", localHash[:8], remoteHash[:8])
		return true
	}
	return false
}

func runDeploy(scriptPath, deployDir string) {
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = deployDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("Deploy failed: %v\n", err)
		return
	}
	log.Println("Deploy completed successfully")
}
