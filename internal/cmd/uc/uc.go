package uc

import (
    "bytes"
    "fmt"
    "github.com/chirino/hawtgo/sh"
    "github.com/chirino/uc/internal/cmd"
    "github.com/chirino/uc/internal/cmd/utils"
    "github.com/chirino/uc/internal/cmd/version"
    "github.com/chirino/uc/internal/pkg/catalog"
    "github.com/chirino/uc/internal/pkg/user"
    "github.com/spf13/cobra"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"
)

func New(o *cmd.Options) (*cobra.Command, error) {

    infoTemp := new(bytes.Buffer)
    debugTemp := new(bytes.Buffer)
    o.InfoLog = infoTemp
    o.DebugLog = debugTemp
    // don't expire the catalog on startup.. user might want to run in offline mode.
    o.CacheExpires = time.Unix(0, 0)

    // in case our binary gets renamed, use that in
    // our help/usage screens.
    use := filepath.Base(os.Args[0])
    if strings.HasPrefix(use, "___") { // looks like an execution from idea.
        use = `uc`
    }
    cacheExpires := "24h"
    verbosity := "info"
    var result = &cobra.Command{
        // BashCompletionFunction: bashCompletionFunction,
        Use:   use,
        Short: `The Kubernetes/OpenShift uber client`,
        //        10        20        30        40        50        60        70        80
        Long: use + ` is an uber client that automatically installs keeps updated Kubernetes and 
OpenShift related command line tools at versions that are best suited to operate 
against the cluster that you are connected to.`,
        Example: `
  ` + use + ` kubectl get pods
  ` + use + ` oc new-project sandbox1
  ` + use + ` kamel run examples/dns.js`,
        TraverseChildren:  true,
        DisableAutoGenTag: true,
        SilenceErrors:     true,
    }
    result.Flags().SetInterspersed(false)
    result.Flags().StringVarP(&o.Kubeconfig, "kubeconfig", "", filepath.Join(user.HomeDir(), ".kube", "config"), "path to the Kubeconfig file")
    result.Flags().StringVarP(&o.Master, "master", "", "", "Master url")
    result.Flags().StringVarP(&cacheExpires, "cache-expires", "", "24h", "Controls when the catalog and command caches expire. One of *duration*|never|now")
    result.Flags().StringVarP(&verbosity, "verbosity", "v", "info", "Sets the verbosity level: One of none|info|debug")

    result.PersistentPreRunE = func(_ *cobra.Command, args []string) error {
        switch strings.ToLower(cacheExpires) {
        case "never":
            o.CacheExpires = time.Unix(0, 0) // way in the past.
        case "now":
            o.CacheExpires = time.Now().Add(10000 * time.Hour) // way in the future..
        default:
            duration, err := time.ParseDuration(cacheExpires)
            if err != nil {
                return fmt.Errorf("invalid flag value --cache-expires '%s': %s", cacheExpires, err)
            }
            o.CacheExpires = time.Now().Add(-duration)
        }

        switch strings.ToLower(verbosity) {
        case "none":
            o.InfoLog = ioutil.Discard
            o.DebugLog = ioutil.Discard
        case "info":
            o.InfoLog = os.Stderr
            o.DebugLog = ioutil.Discard
        case "debug":
            o.InfoLog = os.Stderr
            o.DebugLog = os.Stderr
        default:
            return fmt.Errorf("invalid flag value --verbosity '%s'", verbosity)
        }
        i := infoTemp.Bytes()
        o.InfoLog.Write(i)
        o.DebugLog.Write(debugTemp.Bytes())

        // Now that the --cache-expires is processed, get the catalog again, since it may download a fresh copy due
        // to cache expiration..  Use it to check to see if we need a self update.
        catalog, err := catalog.LoadCatalogIndex(o)
        if err == nil {
            // If the installed uc version is old, get an updated one, and then re-run the CLI against the new version.
            uc := catalog.Commands["uc"]
            if uc != nil && uc.LatestVersion != "" && version.Version != "latest" && uc.LatestVersion != version.Version {
                executable, err := utils.GetExecutable(o, "uc", uc.LatestVersion)
                if err != nil {
                    fmt.Fprintln(o.InfoLog, "self update failed: don't know how get the command executable: ", err)
                } else {
                    err = sh.New().CommandLog(o.DebugLog).CommandLogPrefix("running > ").LineArgs(append([]string{executable}, os.Args[1:]...)...).Exec()
                    if err != nil {
                        fmt.Fprintln(o.InfoLog, "error:", err)
                        os.Exit(3)
                    }
                }
            }
        }

        return nil
    }

    err := addSubcommands(o, result)
    if err != nil {
        return nil, err
    }
    return result, nil
}

func addSubcommands(o *cmd.Options, result *cobra.Command) error {

    subcommands := map[string]*cobra.Command{}
    for _, cmdFactory := range cmd.SubCommandFactories {
        subCommand, err := cmdFactory(o)
        if err != nil {
            return err
        }
        if subCommand != nil {
            subcommands[subCommand.Use] = subCommand
            result.AddCommand(subCommand)
        }
    }

    // options are not yet parsed from the CLI flags, so this basically using
    // --cache-expires never to avoid a doing a network round trip. After parsing this will
    // called against when a sub command is invoked with the right --cache-expires config
    catalog, err := catalog.LoadCatalogIndex(o)
    if err == nil {
        for command, c := range catalog.Commands {

            // the uc catalog entry is to enable self updates, don't add it as a sub command.
            if command == "uc" {
                continue
            }

            subCommand := subcommands[command]
            if subCommand == nil {

                // We dont require a CmdFactory to be created for every command we support.  Only needed if
                // we need more customization than we can do using the data contained in the catalog data.
                // here we setup a command that only exists in the catalog
                subCommand, err = utils.GetCobraCommand(o, command, "latest")
                if err != nil {
                    return err
                }
                subcommands[command] = subCommand
                result.AddCommand(subCommand)
            }
            if c.Short != "" {
                subCommand.Short = c.Short
            }
            if c.Long != "" {
                subCommand.Long = c.Long
            }
            subCommand.Use = command
            subCommand.DisableFlagParsing = true
            subCommand.DisableAutoGenTag = true
        }
    }
    return nil
}
