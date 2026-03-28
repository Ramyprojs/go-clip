package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove goclip from this machine",
	Long:  "Remove the goclip binary, local clipboard history, and configuration from this machine.",
	Args:  cobra.NoArgs,
	RunE:  runUninstallCommand,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstallCommand(cmd *cobra.Command, args []string) error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	resolvedExecutablePath, err := filepath.EvalSymlinks(executablePath)
	if err == nil {
		executablePath = resolvedExecutablePath
	}

	dataDir, err := defaultDataDir()
	if err != nil {
		return err
	}

	confirmed, err := confirmUninstall(cmd, executablePath)
	if err != nil {
		return err
	}

	if !confirmed {
		fmt.Fprintln(cmd.OutOrStdout(), "Uninstall cancelled.")
		return nil
	}

	if err := removeApplicationData(dataDir, appConfig.DBPath); err != nil {
		return err
	}

	if err := scheduleBinaryRemoval(executablePath); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "goclip is uninstalling. The binary will be removed after this process exits.")
	return nil
}

func confirmUninstall(cmd *cobra.Command, executablePath string) (bool, error) {
	fmt.Fprintf(
		cmd.OutOrStdout(),
		"This will remove your goclip binary at %s and delete local history/config files. Continue? [y/N] ",
		executablePath,
	)

	reader := bufio.NewReader(cmd.InOrStdin())
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}

	answer := strings.TrimSpace(strings.ToLower(response))
	return answer == "y" || answer == "yes", nil
}

func defaultDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".goclip"), nil
}

func removeApplicationData(dataDir, databasePath string) error {
	if err := removeCustomDatabase(databasePath, dataDir); err != nil {
		return err
	}

	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("remove goclip data directory: %w", err)
	}

	return nil
}

func removeCustomDatabase(databasePath, dataDir string) error {
	if strings.TrimSpace(databasePath) == "" {
		return nil
	}

	relativePath, err := filepath.Rel(dataDir, databasePath)
	if err != nil {
		return fmt.Errorf("compare database path: %w", err)
	}

	if relativePath == "." || (!strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) && relativePath != "..") {
		return nil
	}

	if err := os.Remove(databasePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove custom database: %w", err)
	}

	return nil
}

func scheduleBinaryRemoval(executablePath string) error {
	if strings.TrimSpace(executablePath) == "" {
		return errors.New("executable path is empty")
	}

	if runtime.GOOS == "windows" {
		return scheduleWindowsBinaryRemoval(executablePath)
	}

	return scheduleUnixBinaryRemoval(executablePath)
}

func scheduleUnixBinaryRemoval(executablePath string) error {
	command := exec.Command("sh", "-c", `sleep 1 && rm -f -- "$1"`, "sh", executablePath)
	if err := command.Start(); err != nil {
		return fmt.Errorf("schedule binary removal: %w", err)
	}

	return nil
}

func scheduleWindowsBinaryRemoval(executablePath string) error {
	command := exec.Command(
		"cmd.exe",
		"/C",
		fmt.Sprintf(`ping 127.0.0.1 -n 2 >NUL && del /f /q %s`, quoteForWindowsCmd(executablePath)),
	)
	if err := command.Start(); err != nil {
		return fmt.Errorf("schedule binary removal: %w", err)
	}

	return nil
}

func quoteForWindowsCmd(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
