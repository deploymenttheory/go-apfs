package discover

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// FormatOutput formats discovery results according to output format
func FormatOutput(response *Response, format string) error {
	switch format {
	case "json":
		return formatJSON(response)
	case "yaml":
		return formatYAML(response)
	case "table":
		return formatTable(response)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// formatTable formats results as a table
func formatTable(response *Response) error {
	if len(response.Files) == 0 {
		fmt.Println("No files found matching the search criteria.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Header
	fmt.Fprintf(w, "PATH\tNAME\tSIZE\tMODIFIED\tTYPE\n")
	fmt.Fprintf(w, "----\t----\t----\t--------\t----\n")

	// Sort files by path for consistent output
	files := make([]FileResult, len(response.Files))
	copy(files, response.Files)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	// Data rows
	for _, file := range files {
		modTime := file.Modified.Format("2006-01-02 15:04")
		if file.Deleted {
			fmt.Fprintf(w, "%s\t%s (deleted)\t%s\t%s\t%s\n",
				file.Path, file.Name, file.FormatSize(), modTime, file.Type)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				file.Path, file.Name, file.FormatSize(), modTime, file.Type)
		}
	}

	// Summary
	fmt.Printf("\n")
	if response.VolumeInfo.Name != "" {
		fmt.Printf("Volume: %s (ID: %d)\n", response.VolumeInfo.Name, response.VolumeInfo.ID)
	}
	fmt.Printf("Found %d files", response.TotalFound)
	if response.Truncated {
		fmt.Printf(" (showing first %d)", len(response.Files))
	}
	fmt.Printf(" in %v\n", response.SearchTime)

	return nil
}

// formatJSON formats results as JSON
func formatJSON(response *Response) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// formatYAML formats results as YAML
func formatYAML(response *Response) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	encoder.SetIndent(2)
	return encoder.Encode(response)
}

// FormatSummary provides a brief summary for verbose output
func FormatSummary(response *Response) string {
	if response.TotalFound == 0 {
		return "No files found"
	}

	summary := fmt.Sprintf("Found %d file", response.TotalFound)
	if response.TotalFound != 1 {
		summary += "s"
	}

	if response.Truncated {
		summary += fmt.Sprintf(" (showing %d)", len(response.Files))
	}

	// Add size breakdown
	var totalSize int64
	sizeClasses := make(map[SizeClass]int)

	for _, file := range response.Files {
		totalSize += file.Size
		sizeClasses[file.GetSizeClass()]++
	}

	summary += fmt.Sprintf(" totaling %s", formatBytes(totalSize))
	summary += fmt.Sprintf(" in %v", response.SearchTime)

	return summary
}

// formatBytes formats byte count as human readable
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
