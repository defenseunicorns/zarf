import TabItem from "@theme/TabItem";
import Tabs from "@theme/Tabs";

# The Zarf CLI

<!-- TODO: @JPERRY This text seems a bit short, What else can we be saying here? -->
<!-- TODO: @JPERRY Is mentioning Cobra actually useful here? -->
<!-- TODO: @JPERRY Should I mention the OS and arch when talking about statically built binaries? -->

Zarf is a command line interface (CLI) tool used to enable software delivery, specifically designed around delivery to disconnected environments. The Zarf tool is a statically built Go binary, meaning once it is built, it can be used anywhere without needing to bring along any other dependencies. The Zarf CLI project is, and always will be, a free to use open-source project on [GitHub](https://github.com/defenseunicorns/zarf).

## Getting the CLI

<!-- TODO: @JPERRY Is it better to link to 'Installing Zarf' or should we repeat the information here? (check w/ Madeline) -->
<!-- TODO: @JPERRY Make sure the 'Installing Zarf' section if fully up to date with all the installation methods -->

There are multiple ways to get the Zarf CLI onto your machine including installing from the Defense Unicorns Homebrew Tap, downloading a prebuilt binary from our GitHub releases, or even building the CLI from scratch yourself. Instructions for all of these methods are provided in the [Installing Zarf](../../3-getting-started.md#installing-zarf) section of the Getting Started guide but if you have Homebrew installed and you want to dive right in, you can install Zarf by copying the commands for your system's OS into a terminal:

<!-- NOTE: The empty line after the '<TabItem ...>' lines are important for the rendering... -->
<Tabs>
<TabItem value="macOS" label="macOS" default>

```bash
brew tap defenseunicorns/tap
brew install zarf
```

</TabItem>

<TabItem value="Linux" label="Linux">

```bash
brew tap defenseunicorns/tap
brew install zarf
```

</TabItem>
</Tabs>

## I have a CLI.. Now What?

<!-- TODO: @JPERRY A lot of this stuff could (and probably should) go in the 'Installing Zarf' section -->

First, lets test to make sure the CLI you have works by running the CLI to get a help message output. Depending on how you installed the CLI, the tabs below will help you figure out how to run the CLI for the first time. If Zarf has been installed properly, you should see a list of all the command options as well as a short description for what each command does.

<details><summary>Expected Help Output</summary>
<p>
The output of the help command should look <b>something</b> like this (CLI flags will also appear at the end of the output):

```text
Zarf eliminates the complexity of air gap software delivery for Kubernetes clusters and cloud native workloads
using a declarative packaging strategy to support DevSecOps in offline and semi-connected environments.

Usage:
  zarf [COMMAND]|[ZARF-PACKAGE]|[ZARF-YAML] [flags]
  zarf [command]

Available Commands:
  completion        Generate the autocompletion script for the specified shell
  connect           Access services or pods deployed in the cluster
  destroy           Tear it all down, we'll miss you Zarf...
  help              Help about any command
  init              Prepares a k8s cluster for the deployment of Zarf packages
  package           Zarf package commands for creating, deploying, and inspecting packages
  prepare           Tools to help prepare assets for packaging
  tools             Collection of additional tools to make airgap easier
  version           Displays the version of the Zarf binary
```

</p>
</details>

<Tabs>
<TabItem value="homebrew" label="Installed via Homebrew" default>

```bash
zarf --help
```

</TabItem>

<TabItem value="custom-install" label="Downloaded from Github">

- If you're not sure where the file was downloaded to, a good default place to look is `~/Downloads`.
- While we only say `zarf` for this example command, the name of the binary is the name of the file you downloaded, which will likely have a different name.

```bash
chmod +x ~/Downloads/zarf   # Make the binary executable
~/Downloaded/zarf --help
```

</TabItem>

<TabItem value="manually-built" label="Manually Built">

- While we only say `zarf` for this example command, depending on your system, you might have to use a different name for the binary like `zarf-mac-intel` or `zarf-mac-apple`

```bash
cd ./path/to/zarf/repo
cd build
./zarf --help
```

</TabItem>

</Tabs>

### Adding The CLI To Your Path

:::note
If you installed Zarf through Homebrew, Zarf will already be on your $PATH. and you can skip this section.
:::

If you want to make your life a little easier, you can put the Zarf CLI on your $PATH. This way, instead of always needing to path to the exact location of the binary, you can just use `zarf` and your computer will automatically find the binary for you to execute. The list of the directories in your PATH can be seen by running `echo $PATH`. As long as you move your CLI to one of those directories you will be able to execute it without having to path to it. One common $PATH you can use is `mv ./path/to/cli/file/zarf /usr/local/bin/zarf`

:::note
Throughout the rest of the demo, we will often be describing commands as `zarf {command}`. This assumes that the CLI is in your path.
:::

## Introduction to Zarf Commands

Zarf has multiple commands to make building, deploying, and maintaining packages easier. Some commands also have multiple sub-commands under them. All of the commands and sub-commands available have a short description of what they do when the `--help` flag is provided. These descriptions get more detailed the further down you go into the command hierarchy. Feel free to explore around the different commands available to get a feel for what Zarf can do.

As stated before, Zarf was built to make deploying applications into disconnected environments easier. To reach this objective, the most common commands that get used are `zarf init`, `zarf package create` and `zarf package deploy`. More detail on all of the commands can be found in the [CLI Commands](./100-cli-commands/zarf.md) section, but short descriptions of the most commonly used commands are provided below. You might notice that all three of these commands operate in some way with what we call a Zarf package. More information about Zarf packages can be found in the next section [Zarf Packages](../2-zarf-packages/1-zarf-packages.md).

### zarf init

<!-- TODO: Find a good place to talk about what the init command is actually doing (there's a lot of special magic sauce going on with that command) -->

`zarf init` is used to prepare a k8s cluster for the deployment of future Zarf packages. The init command uses a specialized 'init-package' to operate. This package can either be located in your current working directory, in the directory where the Zarf CLI binary lives, or be downloaded from the GitHub releases as the command is running. More information about the init-package can be found on the [init-package](../2-zarf-packages/3-the-zarf-init-package.md) page.

### zarf package deploy

<!-- The most common use case (like 99.9% of the time) is deploying onto a k8s cluster.. but that doesn't HAVE to be the case.. How do I write the docs for this then? -->
<!-- TODO: Write some docs (or reddirect to other docs) describing when you would be able to do a `zarf package deploy` before a `zarf init` -->

`zarf package deploy` is used to deploy an already built tar.zst package onto a machine, typically into a k8s cluster. It is usually assumed that the `zarf init` command has already been run on the machine you are deploying to, but there are a few rare cases where this doesn't apply.

## Using a config file to make CLI command flags declarative

The `zarf init`, `zarf package create`, `zarf package deploy` command flags and any global flags can also be set via a config file. The config file Zarf looks for by default is `zarf-config.toml` in the current working directory. Use the command `zarf prepare generate-config` with an optional filename to create a config template for use by Zarf. Note there are several config file formats supported including: `toml`, `json`, `yaml`, `ini` and `props`. For example, to create a template config file with the `my-cool-env` and the yaml format, use the command `zarf prepare generate-config my-cool-env.yaml`.

To use a custom config file, set the `ZARF_CONFIG` environment variable to the path of the config file. For example, to use the `my-cool-env.yaml` config file, set the `ZARF_CONFIG` environment variable to `my-cool-env.yaml`. The `ZARF_CONFIG` environment variable can be set in the shell or in the `.env` file in the current working directory. The `ZARF_CONFIG` environment variable takes precedence over the default config file.

Additionally, any supported config paramater can also be set via env variable using the `ZARF_` prefix. For example, to set the `zarf init` `--storage-class` flag via env variable, set the `ZARF_INIT.STORAGE_CLASS` environment variable. The `ZARF_` environment variable takes precedence over the config file.

Config files set default values, but can still be overwritten by command line flags. For example, if the config file sets the log level to `info` and the command line flag is set to `debug`, the log level will be `debug`. The order of precedence for command line configuration is:

1. Command line flags
2. Environment variables
3. Config file
4. Default values

See the [Config File Example](../../../examples/config-file/README.md) for an example of using a config file.
