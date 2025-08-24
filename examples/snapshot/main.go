package main

import (
	"context"
	"fmt"
	"log"
	"time"

	devento "github.com/devento-ai/sdk-go"
)

func main() {
	// Initialize the client (uses DEVENTO_API_KEY env var)
	client, err := devento.NewClient("")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Use WithSandbox for automatic cleanup
	err = client.WithSandbox(ctx, func(ctx context.Context, box *devento.BoxHandle) error {
		fmt.Printf("Box %s is ready!\n", box.ID())

		// Run initial commands and create a test file
		result, err := box.Run(ctx, `w; echo "Hello from Devento!" | tee /test1; ls -al / | grep test1`, nil)
		if err != nil {
			return fmt.Errorf("failed to run initial command: %w", err)
		}
		fmt.Println("Output:", result.Stdout)
		fmt.Println("Exit code:", result.ExitCode)

		// List existing snapshots (should be empty initially)
		snapshots, err := box.ListSnapshots(ctx)
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}
		fmt.Printf("Existing snapshots: %d\n", len(snapshots))

		// Create a snapshot of the current state
		fmt.Println("\nCreating snapshot...")
		snap, err := box.CreateSnapshot(ctx, "initial-state", "")
		if err != nil {
			return fmt.Errorf("failed to create snapshot: %w", err)
		}
		fmt.Printf("New snapshot: %s - Status: %s\n", snap.ID, snap.Status)

		// Wait for the snapshot to be ready
		fmt.Println("Waiting for snapshot to be ready...")
		err = box.WaitSnapshotReady(ctx, snap.ID, 5*time.Minute, time.Second)
		if err != nil {
			return fmt.Errorf("failed waiting for snapshot: %w", err)
		}
		fmt.Println("Snapshot is ready!")

		// Modify the file
		fmt.Println("\nModifying the file...")
		result2, err := box.Run(ctx, `w; ls -al / | grep test1; cat /test1; echo "new" > /test1`, nil)
		if err != nil {
			return fmt.Errorf("failed to modify file: %w", err)
		}
		fmt.Println("Output:", result2.Stdout)
		fmt.Println("Exit code:", result2.ExitCode)

		// Verify the change
		modifiedContent, err := box.Run(ctx, "cat /test1", nil)
		if err != nil {
			return fmt.Errorf("failed to read modified file: %w", err)
		}
		fmt.Printf("Modified content: %s", modifiedContent.Stdout)

		// Restore from snapshot
		fmt.Printf("\nRestoring snapshot %s...\n", snap.ID)
		restoredSnap, err := box.RestoreSnapshot(ctx, snap.ID)
		if err != nil {
			return fmt.Errorf("failed to restore snapshot: %w", err)
		}
		fmt.Printf("Restore initiated - Status: %s\n", restoredSnap.Status)

		// Wait for the box to be ready after restore
		fmt.Println("Waiting for box to be ready after restore...")
		err = box.WaitUntilReady(ctx)
		if err != nil {
			return fmt.Errorf("failed waiting for box after restore: %w", err)
		}

		// Verify the file is back to original state
		fmt.Println("\nVerifying restore...")
		result3, err := box.Run(ctx, "w; ls -al / | grep test1; cat /test1", nil)
		if err != nil {
			return fmt.Errorf("failed to verify restore: %w", err)
		}
		fmt.Println("Output:", result3.Stdout)
		fmt.Println("Exit code:", result3.ExitCode)

		restoredContent, err := box.Run(ctx, "cat /test1", nil)
		if err != nil {
			return fmt.Errorf("failed to read restored file: %w", err)
		}
		fmt.Printf("Restored content: %s", restoredContent.Stdout)

		// List all snapshots
		finalSnapshots, err := box.ListSnapshots(ctx)
		if err != nil {
			return fmt.Errorf("failed to list final snapshots: %w", err)
		}
		fmt.Printf("\nTotal snapshots: %d\n", len(finalSnapshots))
		for _, s := range finalSnapshots {
			label := s.Label
			if label == "" {
				label = "no label"
			}
			fmt.Printf("  - %s: %s (%s)\n", s.ID, label, s.Status)
		}

		// Clean up: delete the snapshot
		fmt.Printf("\nDeleting snapshot %s...\n", snap.ID)
		deleted, err := box.DeleteSnapshot(ctx, snap.ID)
		if err != nil {
			return fmt.Errorf("failed to delete snapshot: %w", err)
		}
		fmt.Printf("Snapshot deleted - Status: %s\n", deleted.Status)

		return nil
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nSnapshot example completed successfully!")
}
