package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/kim/clockify-cli/internal/api"
	"github.com/kim/clockify-cli/internal/config"
	"github.com/kim/clockify-cli/pkg/clockify"
	"github.com/spf13/cobra"
)

type completionItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type completionNames struct {
	Items []completionItem `json:"items"`
}

const (
	compResourceWorkspaces = "workspaces"
	compResourceProjects   = "projects"
	compResourceClients    = "clients"
	compResourceTags       = "tags"
	compResourceUsers      = "users"
)

func cacheProfileFromCmd(cmd *cobra.Command) string {
	profile, _ := cmd.Flags().GetString("profile")
	return profile
}

func cacheDisabledFromCmd(cmd *cobra.Command) bool {
	// Flag may be missing on some command trees in unit tests.
	if cmd.Flags().Lookup("no-cache") == nil && cmd.PersistentFlags().Lookup("no-cache") == nil {
		return false
	}
	if cmd.Flags().Lookup("no-cache") != nil {
		v, _ := cmd.Flags().GetBool("no-cache")
		return v
	}
	v, _ := cmd.PersistentFlags().GetBool("no-cache")
	return v
}

func workspaceIDFromCmd(cmd *cobra.Command) string {
	profile := cacheProfileFromCmd(cmd)
	cfg, err := config.Load(profile)
	if err != nil || cfg == nil {
		return ""
	}
	return cfg.WorkspaceID
}

func scopedCacheKey(workspaceID, resource string) string {
	if workspaceID == "" {
		return resource + ".json"
	}
	return resource + "." + workspaceID + ".json"
}

func namesCacheKey(workspaceID, resource string) string {
	return "names." + scopedCacheKey(workspaceID, resource)
}

func listCacheKey(workspaceID, resource string) string {
	return "list." + scopedCacheKey(workspaceID, resource)
}

func cacheTTL() time.Duration {
	return config.CacheTTLForEnv()
}

// rememberCompletionNames stores resource names for shell completion (best-effort).
func rememberCompletionNames(cmd *cobra.Command, workspaceID, resource string, items []completionItem) {
	profile := cacheProfileFromCmd(cmd)
	payload := completionNames{Items: items}
	if payload.Items == nil {
		payload.Items = []completionItem{}
	}
	_ = config.WriteTTL(profile, namesCacheKey(workspaceID, resource), payload)
}

// cachedCompletionNames returns names from cache only; never calls the API.
func cachedCompletionNames(cmd *cobra.Command, workspaceID, resource string) []completionItem {
	profile := cacheProfileFromCmd(cmd)
	var payload completionNames
	ok, err := config.ReadTTL(profile, namesCacheKey(workspaceID, resource), &payload, cacheTTL())
	if err != nil || !ok {
		return nil
	}
	return payload.Items
}

// rememberList stores full list payloads for unfiltered list commands.
func rememberList(cmd *cobra.Command, workspaceID, resource string, data any) {
	profile := cacheProfileFromCmd(cmd)
	_ = config.WriteTTL(profile, listCacheKey(workspaceID, resource), data)
}

// cachedList returns a cached list payload when fresh.
func cachedList(cmd *cobra.Command, workspaceID, resource string, out any) bool {
	if cacheDisabledFromCmd(cmd) {
		return false
	}
	profile := cacheProfileFromCmd(cmd)
	ok, err := config.ReadTTL(profile, listCacheKey(workspaceID, resource), out, cacheTTL())
	if err != nil || !ok {
		return false
	}
	return true
}

func namesFromProjects(projects []clockify.Project) []completionItem {
	items := make([]completionItem, 0, len(projects))
	for _, p := range projects {
		items = append(items, completionItem{ID: p.ID, Name: p.Name})
	}
	return items
}

func namesFromClients(clients []clockify.Client) []completionItem {
	items := make([]completionItem, 0, len(clients))
	for _, c := range clients {
		items = append(items, completionItem{ID: c.ID, Name: c.Name})
	}
	return items
}

func namesFromTags(tags []clockify.Tag) []completionItem {
	items := make([]completionItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, completionItem{ID: t.ID, Name: t.Name})
	}
	return items
}

func namesFromUsers(users []clockify.User) []completionItem {
	items := make([]completionItem, 0, len(users))
	for _, u := range users {
		items = append(items, completionItem{ID: u.ID, Name: u.Name})
	}
	return items
}

func namesFromWorkspaces(workspaces []clockify.Workspace) []completionItem {
	items := make([]completionItem, 0, len(workspaces))
	for _, w := range workspaces {
		items = append(items, completionItem{ID: w.ID, Name: w.Name})
	}
	return items
}

// refreshCompletionCache fetches workspaces + current-workspace resources and writes caches.
func refreshCompletionCache(cmd *cobra.Command) error {
	profile := cacheProfileFromCmd(cmd)
	apiKey, region, cfg, err := loadAuthWithConfig(profile)
	if err != nil {
		return err
	}

	ctx := context.Background()
	wsClient := api.NewClient(apiKey, api.BaseURL(region), "")

	workspaces, err := wsClient.ListWorkspaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}
	rememberCompletionNames(cmd, "", compResourceWorkspaces, namesFromWorkspaces(workspaces))
	rememberList(cmd, "", compResourceWorkspaces, workspaces)

	quiet, _ := cmd.Flags().GetBool("quiet")
	counts := map[string]int{compResourceWorkspaces: len(workspaces)}

	if cfg.WorkspaceID == "" {
		if !quiet {
			cmd.Printf("Refreshed %d workspace names. Set a workspace to refresh project/client/tag/user completions.\n", len(workspaces))
		}
		return nil
	}

	client := api.NewClient(apiKey, api.BaseURL(region), cfg.WorkspaceID)
	wsID := cfg.WorkspaceID

	projects, _, err := fetchPaginatedProjects(ctx, client, api.ListProjectsOpts{}, 1, 500, true)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}
	rememberCompletionNames(cmd, wsID, compResourceProjects, namesFromProjects(projects))
	rememberList(cmd, wsID, compResourceProjects, projects)
	counts[compResourceProjects] = len(projects)

	clients, _, err := fetchPaginatedClients(ctx, client, api.ListClientsOpts{}, 1, 500, true)
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}
	rememberCompletionNames(cmd, wsID, compResourceClients, namesFromClients(clients))
	rememberList(cmd, wsID, compResourceClients, clients)
	counts[compResourceClients] = len(clients)

	tags, _, err := fetchPaginatedTags(ctx, client, api.ListTagsOpts{}, 1, 500, true)
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}
	rememberCompletionNames(cmd, wsID, compResourceTags, namesFromTags(tags))
	rememberList(cmd, wsID, compResourceTags, tags)
	counts[compResourceTags] = len(tags)

	users, err := client.ListUsers(ctx, api.ListUsersOpts{})
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}
	rememberCompletionNames(cmd, wsID, compResourceUsers, namesFromUsers(users))
	rememberList(cmd, wsID, compResourceUsers, users)
	counts[compResourceUsers] = len(users)

	if !quiet {
		cmd.Printf("Refreshed completions: %d workspaces, %d projects, %d clients, %d tags, %d users\n",
			counts[compResourceWorkspaces], counts[compResourceProjects], counts[compResourceClients],
			counts[compResourceTags], counts[compResourceUsers])
	}
	return nil
}

// completeCachedNames filters cached names by prefix for shell completion.
func completeCachedNames(cmd *cobra.Command, workspaceID, resource, toComplete string) ([]string, cobra.ShellCompDirective) {
	items := cachedCompletionNames(cmd, workspaceID, resource)
	if len(items) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}
	var out []string
	for _, it := range items {
		if it.Name == "" {
			continue
		}
		if toComplete == "" || containsPrefixFold(it.Name, toComplete) {
			out = append(out, it.Name)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func containsPrefixFold(s, prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	if len(s) < len(prefix) {
		return false
	}
	return equalFoldASCII(s[:len(prefix)], prefix)
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// wrapCompletion adapts a name-filter helper to cobra's CompletionFunc signature.
func wrapCompletion(fn func(*cobra.Command, string) ([]string, cobra.ShellCompDirective)) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return fn(cmd, toComplete)
	}
}

func completeWorkspaceFlag(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeCachedNames(cmd, "", compResourceWorkspaces, toComplete)
}

func completeProfileFlag(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	names, err := config.LoadProfileNames()
	if err != nil || len(names) == 0 {
		return nil, cobra.ShellCompDirectiveError
	}
	var out []string
	for _, n := range names {
		if toComplete == "" || containsPrefixFold(n, toComplete) {
			out = append(out, n)
		}
	}
	return out, cobra.ShellCompDirectiveNoFileComp
}

func completeProjectRef(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeCachedNames(cmd, workspaceIDFromCmd(cmd), compResourceProjects, toComplete)
}

func completeClientRef(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeCachedNames(cmd, workspaceIDFromCmd(cmd), compResourceClients, toComplete)
}

func completeTagRef(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeCachedNames(cmd, workspaceIDFromCmd(cmd), compResourceTags, toComplete)
}

func completeUserRef(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completeCachedNames(cmd, workspaceIDFromCmd(cmd), compResourceUsers, toComplete)
}

// cobraCompletionFuncs are the exported wrappers used with RegisterFlagCompletionFunc.
var (
	completeWorkspaceFlagC = wrapCompletion(completeWorkspaceFlag)
	completeProfileFlagC   = wrapCompletion(completeProfileFlag)
	completeProjectRefC    = wrapCompletion(completeProjectRef)
	completeClientRefC     = wrapCompletion(completeClientRef)
	completeTagRefC        = wrapCompletion(completeTagRef)
	completeUserRefC       = wrapCompletion(completeUserRef)
	// completeTaskRefC is cache-backed project names as a best-effort fallback
	// when no per-project task cache is warm (never blocks on the network).
	completeTaskRefC = wrapCompletion(completeTaskRef)
)

func completeTaskRef(cmd *cobra.Command, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Task names are project-scoped; fall back to project names when task cache
	// is unavailable so tab-complete still suggests something useful.
	return completeCachedNames(cmd, workspaceIDFromCmd(cmd), compResourceProjects, toComplete)
}

func completeSubcommands(cmd *cobra.Command, toComplete string) []string {
	var out []string
	for _, sub := range cmd.Commands() {
		if sub.Hidden || sub.Name() == "help" {
			continue
		}
		name := sub.Name()
		if toComplete == "" || containsPrefixFold(name, toComplete) {
			out = append(out, name)
		}
	}
	return out
}

func completeSubcommandArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeSubcommands(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// wireResourceCompletions attaches cache-backed completions to common resource flags.
func wireResourceCompletions(cmd *cobra.Command, flags map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective)) {
	for flagName, fn := range flags {
		f := cmd.Flags().Lookup(flagName)
		if f == nil {
			continue
		}
		flagFn := fn
		_ = cmd.RegisterFlagCompletionFunc(flagName, wrapCompletion(flagFn))
	}
}

func projectCompletionMap() map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective) {
	return map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective){
		"project":  completeProjectRef,
		"name":     completeProjectRef,
		"new-name": completeProjectRef,
	}
}

func clientCompletionMap() map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective) {
	return map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective){
		"client": completeClientRef,
		"name":   completeClientRef,
	}
}

func tagCompletionMap() map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective) {
	return map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective){
		"tags": completeTagRef,
		"name": completeTagRef,
		"tag":  completeTagRef,
	}
}

func userCompletionMap() map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective) {
	return map[string]func(*cobra.Command, string) ([]string, cobra.ShellCompDirective){
		"user":     completeUserRef,
		"assignee": completeUserRef,
		"email":    completeUserRef,
	}
}
