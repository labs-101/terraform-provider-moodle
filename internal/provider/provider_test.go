package provider

import (
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// providerConfig is built from MOODLE_HOST / MOODLE_TOKEN environment variables.
// Falls back to the local development defaults when the variables are not set.
var providerConfig = func() string {
	host := os.Getenv("MOODLE_HOST")
	if host == "" {
		host = "http://localhost:8080"
	}
	token := os.Getenv("MOODLE_TOKEN")
	if token == "" {
		token = "84306731350159a02855e8dd6dfc1acc"
	}
	return fmt.Sprintf(`
provider "moodle" {
  host  = %q
  token = %q
}
`, host, token)
}()

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"moodle": providerserver.NewProtocol6WithError(New("test")()),
	}
)
