[![MIT License][license-shield]][license-url]

# Mantis ![Version][version-tag]

#### About

A simple tool that helps with golang development by restarting your application when modifications are detected. Inspired from [Nodemon](https://www.npmjs.com/package/nodemon)

#### Installation

```
go install github.com/nn-advith/mantis@latest
```

For specific version (eg: v1.0.0)

```
go install github.com/nn-advith/mantis@v1.0.0
```

Mantis will be installed on your system.

#### Usage

Bare minimum command

```
mantis -f sample.go
```

This starts up sample.go and monitors for any changes. The output and errors are redirected from the process to Stdout and Stderr.

For help, use `-h` flag

```
mantis -h
```

Additional flags supported by Mantis cli:

```
mantis -v   -   version info
       -d   -   delay in milliseconds before starting the application
       -a   -   arguments to be passed to the application
       -e   -   environment variables to be passed to the application
```

#### Config files

Mantis supports the usage of config files to simplify command execution. There are two types; global config and local config. A global config file is created on your system, which contains a minimal config entry. Users can create a local config file in the directory of your program to save the flags for simpler command execution.

The priority is `Inline values > Local Config file > Global Config file`

Config file keys:

```
extensions  -   The file extensions that are to be monitored seperated by comma.
                All files matching this willbe monitored, unless ignored
ignore      -   Comma seperated list of files/directories to be ignored while checking for modifications
delay       -   The delay in milliseconds before starting the processes.
env         -   The environment variables to be passed into the applicatio separated by comma.
args        -   The arguments to be passed into the applicatio separated by comma.
```

##### Global Config

Minimal config file, `mantis.json` for global config. This file can be modified as per your needs, but note this changes will be used for all Mantis cli executions. For a finer grain of control, use Local Config.

```
{
    "extensions": ".go",
    "ignore": "",
    "delay": "0",
    "env": "",
    "args": ""
}
```

On Windows, global config file path is `%APPDATA%/mantis/mantis.json`
On Linux, global config file path is `$HOME/.config/mantis/mantis.json`

##### Local Config

In the directory of the program that you need to monitor, create a config file `mantis.json` and add your config file content. For example:

```
{
    "extensions": ".go,.mod",
    "ignore": "somedir/",
    "delay": "0",
    "env": "key=val",
    "args": "arg1,arg2"
}
```

Now you can execute `mantis -f .` in the program directory and the config file will be used to read all the flags. Note that adding any other flag in the cli invocation will override it's corresponding value form the config. For instance, `mantis -f . -d 1000` will cause a 1 second delay.

[license-shield]: https://img.shields.io/badge/LICENSE-MIT-green?style=flat&labelColor=%232a2a2a&color=%2365ff8a
[license-url]: https://github.com/nn-advith/mantis/blob/main/LICENSE
[windows-installer-url]: https://github.com/nn-advith/nyx/releases/download/v1.0.0/Nyx-Setup-1.0.0.exe
[windows-installer-icon]: https://img.shields.io/badge/windows-installer-blue?style=for-the-badge
[version-tag]: https://img.shields.io/badge/v-1.0.0-green?style=flat&labelColor=%232a2a2a&color=%2365ff8a
