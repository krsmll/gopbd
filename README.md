# goub

## osu! User Beatmap Downloader

## Installation:

- Install Go from [the official website](https://go.dev/).
- Clone this project using [Git](https://git-scm.com/) or download source code directly from Github.
- Open `cmd` or any other command-line interface in the project directory.
- Run `go get` to download necessary dependencies.
- Run `go build` to build an executable file.
- Place that executable file anywhere on your computer.
- Add the path to that file in PATH environment
  variable. [Windows tutorial here.](https://stackoverflow.com/questions/44272416/how-to-add-a-folder-to-path-environment-variable-in-windows-10-with-screensho)
- Restart your command-line interface.
- Run `goub --help`.

Executables will be added to Releases when I am absolutely sure it's done.

[About being flagged as virus.](https://go.dev/doc/faq#virus)

## How to Use:

You can access the help command using `--help` or `-h`.

### Configuration file generation

Configuration file generation is fairly straightforward. The following command will create a configuration file in your
home directory.

```console
goub generate_config --client_id YOUR_CLIENT_ID --client_secret YOUR_CLIENT_SECRET
```

Obviously, replace `YOUR_CLIENT_ID` and `YOUR_CLIENT_ID_HERE` with your **osu! API v2** credentials.

### Downloading beatmaps

The following downloads all Ryuusei Aika's ranked and favorite beatmapsets.
```console
goub download -u 7777875 --ranked --favorite
```

`goub --help` for more.
