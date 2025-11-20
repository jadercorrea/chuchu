package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getSetupPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".chuchu", "setup.yaml"), nil
}

type BackendProfile struct {
	Name        string
	AgentModels map[string]string
}

func ListBackendProfiles(backendName string) ([]string, error) {
	setupPath, err := getSetupPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(setupPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	profiles := []string{}
	scanner := bufio.NewScanner(file)
	inTargetBackend := false
	inProfiles := false
	hasBackendAgentModels := false

	for scanner.Scan() {
		line := scanner.Text()
		indent := countSpaces(line)

		if indent == 4 && strings.HasPrefix(strings.TrimSpace(line), backendName+":") {
			inTargetBackend = true
		} else if indent == 4 && inTargetBackend && strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), " ") {
			break
		}

		if inTargetBackend {
			if indent == 8 && strings.HasPrefix(strings.TrimSpace(line), "agent_models:") {
				hasBackendAgentModels = true
			}
			if indent == 8 && strings.HasPrefix(strings.TrimSpace(line), "profiles:") {
				inProfiles = true
			} else if inProfiles && indent == 12 && strings.Contains(line, ":") {
				profileName := strings.TrimSpace(strings.Split(line, ":")[0])
				profiles = append(profiles, profileName)
			} else if inProfiles && indent < 12 {
				inProfiles = false
			}
		}
	}

	if hasBackendAgentModels {
		profiles = append([]string{"default"}, profiles...)
	}

	if len(profiles) == 0 {
		profiles = append(profiles, "default")
	}

	return profiles, scanner.Err()
}

func GetBackendProfile(backendName, profileName string) (*BackendProfile, error) {
	setupPath, err := getSetupPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(setupPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	profile := &BackendProfile{
		Name:        profileName,
		AgentModels: make(map[string]string),
	}

	scanner := bufio.NewScanner(file)
	inTargetBackend := false
	inProfiles := false
	inTargetProfile := false
	inAgentModels := false
	inBackendAgentModels := false

	for scanner.Scan() {
		line := scanner.Text()
		indent := countSpaces(line)
		trimmed := strings.TrimSpace(line)

		if indent == 4 && strings.HasPrefix(trimmed, backendName+":") {
			inTargetBackend = true
		} else if indent == 4 && inTargetBackend && trimmed != "" {
			break
		}

		if !inTargetBackend {
			continue
		}

		if indent == 8 && strings.HasPrefix(trimmed, "profiles:") {
			inProfiles = true
			inBackendAgentModels = false
		} else if indent == 8 && strings.HasPrefix(trimmed, "agent_models:") && !inProfiles {
			inBackendAgentModels = true
		} else if inProfiles && indent == 12 && strings.HasPrefix(trimmed, profileName+":") {
			inTargetProfile = true
			inBackendAgentModels = false
		} else if inTargetProfile && indent == 16 && strings.HasPrefix(trimmed, "agent_models:") {
			inAgentModels = true
		} else if inAgentModels && indent == 20 && strings.Contains(line, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				agent := strings.TrimSpace(parts[0])
				model := strings.TrimSpace(parts[1])
				profile.AgentModels[agent] = model
			}
		} else if inAgentModels && indent < 20 {
			inAgentModels = false
		} else if inBackendAgentModels && profileName == "default" && indent == 12 && strings.Contains(line, ":") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				agent := strings.TrimSpace(parts[0])
				model := strings.TrimSpace(parts[1])
				profile.AgentModels[agent] = model
			}
		}
	}

	return profile, scanner.Err()
}

func CreateBackendProfile(backendName, profileName string) error {
	setupPath, err := getSetupPath()
	if err != nil {
		return err
	}

	file, err := os.Open(setupPath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	inTargetBackend := false
	profilesLineIdx := -1
	backendEndIdx := -1

	for i, line := range lines {
		indent := countSpaces(line)
		trimmed := strings.TrimSpace(line)

		if indent == 4 && strings.HasPrefix(trimmed, backendName+":") {
			inTargetBackend = true
		} else if indent == 4 && inTargetBackend && trimmed != "" {
			backendEndIdx = i
			break
		}

		if inTargetBackend && indent == 8 && strings.HasPrefix(trimmed, "profiles:") {
			profilesLineIdx = i
		}
	}

	if !inTargetBackend {
		return fmt.Errorf("backend %s not found in setup.yaml", backendName)
	}

	newProfileLines := []string{
		strings.Repeat(" ", 12) + profileName + ":",
		strings.Repeat(" ", 16) + "agent_models:",
		strings.Repeat(" ", 20) + "router: ",
		strings.Repeat(" ", 20) + "query: ",
		strings.Repeat(" ", 20) + "editor: ",
		strings.Repeat(" ", 20) + "research: ",
	}

	var newLines []string
	if profilesLineIdx >= 0 {
		newLines = append(lines[:profilesLineIdx+1], newProfileLines...)
		newLines = append(newLines, lines[profilesLineIdx+1:]...)
	} else {
		insertIdx := backendEndIdx
		if insertIdx < 0 {
			insertIdx = len(lines)
		}
		
		profilesHeader := strings.Repeat(" ", 8) + "profiles:"
		newLines = append(lines[:insertIdx], profilesHeader)
		newLines = append(newLines, newProfileLines...)
		newLines = append(newLines, lines[insertIdx:]...)
	}

	output := strings.Join(newLines, "\n") + "\n"
	return os.WriteFile(setupPath, []byte(output), 0644)
}

func SetProfileAgentModel(backendName, profileName, agent, model string) error {
	setupPath, err := getSetupPath()
	if err != nil {
		return err
	}

	file, err := os.Open(setupPath)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		return err
	}

	inTargetBackend := false
	inProfiles := false
	inTargetProfile := false
	inAgentModels := false
	inBackendAgentModels := false
	agentLineIdx := -1

	for i, line := range lines {
		indent := countSpaces(line)
		trimmed := strings.TrimSpace(line)

		if indent == 4 && strings.HasPrefix(trimmed, backendName+":") {
			inTargetBackend = true
		} else if indent == 4 && inTargetBackend && trimmed != "" {
			break
		}

		if !inTargetBackend {
			continue
		}

		if indent == 8 && strings.HasPrefix(trimmed, "profiles:") {
			inProfiles = true
			inBackendAgentModels = false
		} else if indent == 8 && strings.HasPrefix(trimmed, "agent_models:") && !inProfiles {
			inBackendAgentModels = true
		} else if inProfiles && indent == 12 && strings.HasPrefix(trimmed, profileName+":") {
			inTargetProfile = true
		} else if inTargetProfile && indent == 16 && strings.HasPrefix(trimmed, "agent_models:") {
			inAgentModels = true
		} else if inAgentModels && indent == 20 && strings.HasPrefix(trimmed, agent+":") {
			agentLineIdx = i
			break
		} else if inBackendAgentModels && profileName == "default" && indent == 12 && strings.HasPrefix(trimmed, agent+":") {
			agentLineIdx = i
			break
		}
	}

	if agentLineIdx < 0 {
		return fmt.Errorf("agent %s not found in profile %s/%s", agent, backendName, profileName)
	}

	indent := countSpaces(lines[agentLineIdx])
	lines[agentLineIdx] = strings.Repeat(" ", indent) + agent + ": " + model

	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(setupPath, []byte(output), 0644)
}

func countSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}
