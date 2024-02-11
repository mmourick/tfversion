package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// IsAlreadyDownloaded checks if the given Terraform version is already downloaded and unzipped.
func IsAlreadyDownloaded(version string) bool {
	downloadLocation := getDownloadLocation()
	filePath := filepath.Join(downloadLocation, version, TerraformBinaryName)
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func getDownloadLocation() string {
	user, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home directory: %s", err)
	}
	return filepath.Join(user, DownloadDir)
}

func ensureDownloadDirectoryExists(downloadLocation string) {
	if _, err := os.Stat(downloadLocation); os.IsNotExist(err) {
		err := os.Mkdir(downloadLocation, 0755)
		if err != nil {
			fmt.Printf("error creating download directory: %s", err)
		}
	}
}

// Download downloads the Terraform release zip file for the given version, OS and architecture.
func Download(version, goos, goarch string) (string, error) {
	downloadLocation := getDownloadLocation()
	ensureDownloadDirectoryExists(downloadLocation)

	// Construct the download URL based on the version and the OS and architecture.
	downloadURL := fmt.Sprintf("%s/%s/terraform_%s_%s_%s.zip", TerraformReleasesUrl, version, version, goos, goarch)

	var err error
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		if err = downloadWithRetry(downloadURL, downloadLocation, version, goos, goarch); err == nil {
			fmt.Printf("Terraform %s downloaded successfully\n", version)
			// Return the path to the downloaded file.
			return fmt.Sprintf("%s/terraform_%s_%s_%s.zip", downloadLocation, version, goos, goarch), nil
		}

		fmt.Printf("Attempt %d failed: %s\n", attempt, err)
		time.Sleep(time.Second * RetryTimeInSeconds) // sleep before retrying.
	}

	// If we got here, we failed to download Terraform after MaxRetries attempts.
	return "", fmt.Errorf("failed to download Terraform after %d attempts: %s", MaxRetries, err)
}

func downloadWithRetry(downloadURL, downloadLocation, version, goos, goarch string) error {
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download Terraform: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Terraform: %s", resp.Status)
	}

	filePath := filepath.Join(downloadLocation, fmt.Sprintf("terraform_%s_%s_%s.zip", version, goos, goarch))
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: %s", err)
	}

	return nil
}

func DeleteDownloadedRelease(zipFile string) error {
	err := os.Remove(zipFile)
	if err != nil {
		return fmt.Errorf("failed to delete Terraform release: %s", err)
	}
	return nil
}