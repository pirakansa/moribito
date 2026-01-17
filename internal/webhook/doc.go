// Package webhook provides event handlers for GitHub webhook events.
// Each handler parses the event payload and enqueues a background job
// for processing.
package webhook
