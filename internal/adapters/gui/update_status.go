//go:build gui

// Maps updater outcomes to safe explanatory HUD status.
// Uses package-local contracts and Go standard-library types.
package gui

type UpdateStatusCode string

const (
	UpdateStatusChecking           UpdateStatusCode = "checking"
	UpdateStatusUpToDate           UpdateStatusCode = "up_to_date"
	UpdateStatusAvailable          UpdateStatusCode = "update_available"
	UpdateStatusUnableNoChecker    UpdateStatusCode = "unable_no_checker"
	UpdateStatusUnableBridge       UpdateStatusCode = "unable_bridge"
	UpdateStatusUnableNetwork      UpdateStatusCode = "unable_network"
	UpdateStatusUnableHTTP         UpdateStatusCode = "unable_http"
	UpdateStatusUnableDownload     UpdateStatusCode = "unable_download"
	UpdateStatusUnableVerification UpdateStatusCode = "unable_verification"
	UpdateStatusUnablePackage      UpdateStatusCode = "unable_package"
	UpdateStatusUnableSigner       UpdateStatusCode = "unable_signer"
	UpdateStatusNoNewerUpdate      UpdateStatusCode = "no_newer_update"
	UpdateStatusPermissionNeeded   UpdateStatusCode = "permission_needed"
	UpdateStatusInstallerOpened    UpdateStatusCode = "installer_opened"
	UpdateStatusUnableInstall      UpdateStatusCode = "unable_install"
)

type updateStatus struct {
	Label  string
	Detail string
	Reason UpdateStatusCode
}

func updateStatusFor(code UpdateStatusCode) updateStatus {
	switch code {
	case UpdateStatusChecking:
		return updateStatus{Label: "Checking...", Detail: "Contacting the configured update service.", Reason: code}
	case UpdateStatusUpToDate:
		return updateStatus{Label: "Up To Date!", Detail: "The downloaded APK matches the installed APK.", Reason: code}
	case UpdateStatusAvailable:
		return updateStatus{Label: "Update available", Detail: "A verified update is ready for installation.", Reason: code}
	case UpdateStatusPermissionNeeded:
		return updateStatus{Label: "Permission needed", Detail: "Allow Apparat to install updates, then check again.", Reason: code}
	case UpdateStatusInstallerOpened:
		return updateStatus{Label: "Installer opened", Detail: "Confirm the update in the system installer.", Reason: code}
	case UpdateStatusUnableNetwork:
		return updateStatus{Label: "Download failed", Detail: "The update could not be downloaded or saved. Check your connection and available storage, then retry. No changes were installed.", Reason: code}
	case "unable_timeout":
		return updateStatus{Label: "Download timed out", Detail: "The update download stalled or exceeded its time limit. Retry on a stable connection. No changes were installed.", Reason: code}
	case "unable_size":
		return updateStatus{Label: "Update rejected", Detail: "The download exceeds the supported APK size. No changes were installed.", Reason: code}
	case "update_cancelled":
		return updateStatus{Label: "Download cancelled", Detail: "The update download was cancelled. No changes were installed.", Reason: code}
	case UpdateStatusUnableHTTP:
		return updateStatus{Label: "Unable to check for updates", Detail: "The update service returned an unexpected response. Try again later.", Reason: code}
	case UpdateStatusUnableDownload:
		return updateStatus{Label: "Unable to check for updates", Detail: "The update could not be downloaded. No changes were installed.", Reason: code}
	case UpdateStatusUnablePackage:
		return updateStatus{Label: "Update rejected", Detail: "The download is not a valid Apparat package. No changes were installed.", Reason: code}
	case UpdateStatusUnableSigner:
		return updateStatus{Label: "Update rejected", Detail: "The download uses a different signing key from this installation. A compatible release is required; keep your installed app.", Reason: code}
	case UpdateStatusNoNewerUpdate:
		return updateStatus{Label: "No newer update", Detail: "The published APK has the same or an older version number. Your installed app was kept.", Reason: code}
	case UpdateStatusUnableVerification:
		return updateStatus{Label: "Unable to check for updates", Detail: "The downloaded update could not be verified. No changes were installed.", Reason: code}
	case UpdateStatusUnableBridge:
		return updateStatus{Label: "Unable to check for updates", Detail: "The Android update service is not ready. Try again after the app finishes starting.", Reason: code}
	case UpdateStatusUnableInstall:
		return updateStatus{Label: "Unable to install update", Detail: "The verified update could not be opened by the system installer.", Reason: code}
	case UpdateStatusUnableNoChecker:
		fallthrough
	default:
		return updateStatus{Label: "Unable to check for updates", Detail: "No update checker is configured in this build.", Reason: UpdateStatusUnableNoChecker}
	}
}
