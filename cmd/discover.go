package cmd

import (
	"github.com/spf13/cobra"

	"github.com/deploymenttheory/go-apfs/pkg/app"
	"github.com/deploymenttheory/go-apfs/pkg/app/discover"
)

var (
	// Volume/snapshot selection (discover command only)
	discoverVolumeID   uint64
	discoverVolumeName string
	discoverSnapshot   string

	// File matching criteria
	namePattern   string
	nameRegex     string
	extensions    []string
	caseSensitive bool

	// Size criteria
	minSize string
	maxSize string

	// Date criteria
	modifiedAfter  string
	modifiedBefore string

	// Content search
	contentSearch  string
	includeDeleted bool
	maxResults     int
)

var discoverCmd = &cobra.Command{
	Use:   "discover [container-path]",
	Short: "Find files within APFS containers by name, size, date, or content",
	Long: `Search for files within APFS containers using various criteria.

Examples:
  # Find all PDF files in volume "Macintosh HD"
  go-apfs discover backup.dmg --volume-name "Macintosh HD" --ext pdf

  # Find files with "password" in name
  go-apfs discover /dev/disk2 --name "*password*"

  # Find large files in a specific snapshot
  go-apfs discover backup.dmg --snapshot "Daily-2024-01-15" --min-size 100MB

  # Search file contents for specific text
  go-apfs discover backup.dmg --content "secret" --ext txt,log`,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDiscover(args[0])
	},
}

func init() {
	rootCmd.AddCommand(discoverCmd)

	// Volume/snapshot selection
	discoverCmd.Flags().Uint64Var(&discoverVolumeID, "volume-id", 0, "volume ID to search")
	discoverCmd.Flags().StringVar(&discoverVolumeName, "volume-name", "", "volume name to search")
	discoverCmd.Flags().StringVar(&discoverSnapshot, "snapshot", "", "snapshot to search")

	// File matching
	discoverCmd.Flags().StringVarP(&namePattern, "name", "n", "", "filename pattern (wildcards: *, ?)")
	discoverCmd.Flags().StringVar(&nameRegex, "regex", "", "filename regex pattern")
	discoverCmd.Flags().StringSliceVar(&extensions, "ext", nil, "file extensions (pdf,jpg,txt)")
	discoverCmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "case-sensitive matching")

	// Size filtering
	discoverCmd.Flags().StringVar(&minSize, "min-size", "", "minimum file size (10MB, 1GB)")
	discoverCmd.Flags().StringVar(&maxSize, "max-size", "", "maximum file size (100MB, 2GB)")

	// Date filtering
	discoverCmd.Flags().StringVar(&modifiedAfter, "after", "", "modified after (YYYY-MM-DD)")
	discoverCmd.Flags().StringVar(&modifiedBefore, "before", "", "modified before (YYYY-MM-DD)")

	// Content search
	discoverCmd.Flags().StringVarP(&contentSearch, "content", "c", "", "search text within files")
	discoverCmd.Flags().BoolVar(&includeDeleted, "deleted", false, "include deleted files")
	discoverCmd.Flags().IntVar(&maxResults, "limit", 1000, "maximum results")

	// Mutual exclusions
	discoverCmd.MarkFlagsMutuallyExclusive("volume-id", "volume-name")
	discoverCmd.MarkFlagsMutuallyExclusive("name", "regex")
}

func runDiscover(containerPath string) error {
	// Create application context
	ctx := app.NewContext()
	ctx.OutputFormat = GetOutputFormat()
	ctx.Verbose = GetVerbose()
	ctx.Quiet = GetQuiet()

	// Create discovery request
	request := &discover.Request{
		ContainerPath: containerPath,
		Target: app.VolumeTarget{
			VolumeID:   discoverVolumeID,
			VolumeName: discoverVolumeName,
			Snapshot:   discoverSnapshot,
		},
		NamePattern:    namePattern,
		NameRegex:      nameRegex,
		Extensions:     extensions,
		CaseSensitive:  caseSensitive,
		MinSize:        minSize,
		MaxSize:        maxSize,
		ModifiedAfter:  modifiedAfter,
		ModifiedBefore: modifiedBefore,
		ContentSearch:  contentSearch,
		IncludeDeleted: includeDeleted,
		MaxResults:     maxResults,
	}

	// Handle the request through application layer
	response, err := discover.Handle(ctx, request)
	if err != nil {
		return err
	}

	// Format and display results
	return discover.FormatOutput(response, ctx.OutputFormat)
}
