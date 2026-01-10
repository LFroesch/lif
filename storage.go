package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

func initializeReference() []ReferenceItem {
	return []ReferenceItem{
		// Git commands
		{ID: 1, Lang: "git", Command: "git clone", Usage: "git clone <url>", Example: "git clone https://github.com/user/repo.git", Meaning: "clone a repository"},
		{ID: 2, Lang: "git", Command: "git status", Usage: "git status", Example: "git status", Meaning: "show working tree status"},
		{ID: 3, Lang: "git", Command: "git add", Usage: "git add <file>", Example: "git add .", Meaning: "stage changes for commit"},
		{ID: 4, Lang: "git", Command: "git commit", Usage: "git commit -m \"message\"", Example: "git commit -m \"fix bug\"", Meaning: "commit staged changes"},
		{ID: 5, Lang: "git", Command: "git push", Usage: "git push <remote> <branch>", Example: "git push origin main", Meaning: "push commits to remote"},
		{ID: 6, Lang: "git", Command: "git pull", Usage: "git pull <remote> <branch>", Example: "git pull origin main", Meaning: "fetch and merge remote changes"},
		{ID: 7, Lang: "git", Command: "git branch", Usage: "git branch <name>", Example: "git branch feature-x", Meaning: "create or list branches"},
		{ID: 8, Lang: "git", Command: "git checkout", Usage: "git checkout <branch>", Example: "git checkout develop", Meaning: "switch branches"},
		{ID: 9, Lang: "git", Command: "git merge", Usage: "git merge <branch>", Example: "git merge feature-x", Meaning: "merge branch into current"},
		{ID: 10, Lang: "git", Command: "git log", Usage: "git log", Example: "git log --oneline", Meaning: "show commit history"},
		{ID: 11, Lang: "git", Command: "git diff", Usage: "git diff", Example: "git diff HEAD~1", Meaning: "show changes between commits"},
		{ID: 12, Lang: "git", Command: "git stash", Usage: "git stash", Example: "git stash pop", Meaning: "temporarily save changes"},
		{ID: 13, Lang: "git", Command: "git rebase", Usage: "git rebase <branch>", Example: "git rebase main", Meaning: "reapply commits on top of another base"},
		{ID: 14, Lang: "git", Command: "git reset", Usage: "git reset <file>", Example: "git reset HEAD~1", Meaning: "undo commits or unstage changes"},

		// Docker commands
		{ID: 15, Lang: "docker", Command: "docker build", Usage: "docker build -t <name> .", Example: "docker build -t myapp .", Meaning: "build image from dockerfile"},
		{ID: 16, Lang: "docker", Command: "docker run", Usage: "docker run <image>", Example: "docker run -p 8080:80 nginx", Meaning: "run a container from image"},
		{ID: 17, Lang: "docker", Command: "docker ps", Usage: "docker ps", Example: "docker ps -a", Meaning: "list running containers"},
		{ID: 18, Lang: "docker", Command: "docker stop", Usage: "docker stop <container>", Example: "docker stop myapp", Meaning: "stop running container"},
		{ID: 19, Lang: "docker", Command: "docker rm", Usage: "docker rm <container>", Example: "docker rm myapp", Meaning: "remove stopped container"},
		{ID: 20, Lang: "docker", Command: "docker images", Usage: "docker images", Example: "docker images", Meaning: "list docker images"},
		{ID: 21, Lang: "docker", Command: "docker exec", Usage: "docker exec -it <container> <cmd>", Example: "docker exec -it myapp bash", Meaning: "execute command in container"},
		{ID: 22, Lang: "docker", Command: "docker logs", Usage: "docker logs <container>", Example: "docker logs -f myapp", Meaning: "view container logs"},
		{ID: 23, Lang: "docker", Command: "docker compose up", Usage: "docker compose up", Example: "docker compose up -d", Meaning: "start services from compose file"},
		{ID: 24, Lang: "docker", Command: "docker compose down", Usage: "docker compose down", Example: "docker compose down", Meaning: "stop and remove containers"},

		// npm commands
		{ID: 25, Lang: "npm", Command: "npm init", Usage: "npm init", Example: "npm init -y", Meaning: "initialize new package"},
		{ID: 26, Lang: "npm", Command: "npm install", Usage: "npm install <package>", Example: "npm install express", Meaning: "install package"},
		{ID: 27, Lang: "npm", Command: "npm run", Usage: "npm run <script>", Example: "npm run dev", Meaning: "run package.json script"},
		{ID: 28, Lang: "npm", Command: "npm test", Usage: "npm test", Example: "npm test", Meaning: "run tests"},
		{ID: 29, Lang: "npm", Command: "npm update", Usage: "npm update", Example: "npm update", Meaning: "update packages"},
		{ID: 30, Lang: "npm", Command: "npm uninstall", Usage: "npm uninstall <package>", Example: "npm uninstall lodash", Meaning: "remove package"},

		// curl commands
		{ID: 31, Lang: "curl", Command: "curl GET", Usage: "curl <url>", Example: "curl https://api.example.com", Meaning: "make http get request"},
		{ID: 32, Lang: "curl", Command: "curl POST", Usage: "curl -X POST -d \"data\" <url>", Example: "curl -X POST -d '{\"key\":\"value\"}' api.com", Meaning: "make http post request"},
		{ID: 33, Lang: "curl", Command: "curl headers", Usage: "curl -H \"Header: value\" <url>", Example: "curl -H \"Authorization: Bearer token\" api.com", Meaning: "send request with headers"},
		{ID: 34, Lang: "curl", Command: "curl download", Usage: "curl -O <url>", Example: "curl -O https://example.com/file.zip", Meaning: "download file"},

		// Linux/bash commands
		{ID: 35, Lang: "bash", Command: "grep", Usage: "grep <pattern> <file>", Example: "grep \"error\" log.txt", Meaning: "search for pattern in file"},
		{ID: 36, Lang: "bash", Command: "find", Usage: "find <path> -name <pattern>", Example: "find . -name \"*.js\"", Meaning: "find files by pattern"},
		{ID: 37, Lang: "bash", Command: "chmod", Usage: "chmod <permissions> <file>", Example: "chmod +x script.sh", Meaning: "change file permissions"},
		{ID: 38, Lang: "bash", Command: "chown", Usage: "chown <user>:<group> <file>", Example: "chown user:group file.txt", Meaning: "change file ownership"},
		{ID: 39, Lang: "bash", Command: "tar", Usage: "tar -czf <archive> <files>", Example: "tar -czf backup.tar.gz folder/", Meaning: "create compressed archive"},
		{ID: 40, Lang: "bash", Command: "untar", Usage: "tar -xzf <archive>", Example: "tar -xzf backup.tar.gz", Meaning: "extract compressed archive"},
		{ID: 41, Lang: "bash", Command: "ssh", Usage: "ssh <user>@<host>", Example: "ssh user@192.168.1.1", Meaning: "connect to remote server"},
		{ID: 42, Lang: "bash", Command: "scp", Usage: "scp <source> <user>@<host>:<dest>", Example: "scp file.txt user@server:/path/", Meaning: "copy files over ssh"},
		{ID: 43, Lang: "bash", Command: "ps", Usage: "ps aux", Example: "ps aux | grep node", Meaning: "list running processes"},
		{ID: 44, Lang: "bash", Command: "kill", Usage: "kill <pid>", Example: "kill -9 1234", Meaning: "terminate process"},
		{ID: 45, Lang: "bash", Command: "systemctl", Usage: "systemctl <action> <service>", Example: "systemctl restart nginx", Meaning: "manage system services"},
		{ID: 46, Lang: "bash", Command: "tail", Usage: "tail -f <file>", Example: "tail -f /var/log/syslog", Meaning: "follow file updates"},
		{ID: 47, Lang: "bash", Command: "sed", Usage: "sed 's/old/new/g' <file>", Example: "sed 's/foo/bar/g' file.txt", Meaning: "stream editor for text"},
		{ID: 48, Lang: "bash", Command: "awk", Usage: "awk '{print $1}' <file>", Example: "awk '{print $2}' data.txt", Meaning: "pattern scanning and processing"},

		// Go commands
		{ID: 49, Lang: "go", Command: "go run", Usage: "go run <file>", Example: "go run main.go", Meaning: "compile and run go program"},
		{ID: 50, Lang: "go", Command: "go build", Usage: "go build", Example: "go build -o app", Meaning: "compile go program"},
		{ID: 51, Lang: "go", Command: "go test", Usage: "go test", Example: "go test ./...", Meaning: "run tests"},
		{ID: 52, Lang: "go", Command: "go mod init", Usage: "go mod init <module>", Example: "go mod init github.com/user/repo", Meaning: "initialize go module"},
		{ID: 53, Lang: "go", Command: "go get", Usage: "go get <package>", Example: "go get github.com/pkg/errors", Meaning: "download and install package"},
	}
}

func loadData() AppData {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(configDir, "lif", "config.json")

	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(configPath), 0755)

	data := AppData{
		Dailies:      []Daily{},
		RollingTodos: []RollingTodo{},
		Reminders:    []Reminder{},
		Reference:    initializeReference(),
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		saveData(data)
		return data
	}

	file, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(file, &data); err != nil {
		log.Printf("Warning: failed to parse config: %v. Using defaults.", err)
		return data
	}

	// Initialize reminders that need parsing
	for i := range data.Reminders {
		reminder := &data.Reminders[i]
		if reminder.TargetTime.IsZero() && reminder.AlarmOrCountdown != "" {
			if targetTime, isCountdown := parseCountdown(reminder.AlarmOrCountdown); isCountdown {
				reminder.TargetTime = targetTime
				reminder.IsCountdown = true
				reminder.Status = "active"
			} else if targetTime, isAlarm := parseAlarmTime(reminder.AlarmOrCountdown); isAlarm {
				reminder.TargetTime = targetTime
				reminder.IsCountdown = false
				reminder.Status = "active"
			}
		}
	}

	return data
}

func saveData(data AppData) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	configPath := filepath.Join(configDir, "lif", "config.json")

	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(configPath, file, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
