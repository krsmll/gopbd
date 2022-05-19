# goub

## osu! User Beatmap Downloader

[About being flagged as virus.](https://go.dev/doc/faq#virus)

## Installation

There are two ways:

### Precompiled Executable from Releases 

- Download a precompiled executable from Releases.
- Place it anywhere on your computer.
- Run the executable using your command-line interface of choice.


### Compile Yourself

- Install Go from [the official website](https://go.dev/).
- Clone this project using [Git](https://git-scm.com/) or download source code directly from Github.
- Open your command-line interface of choice (e.g. `cmd` for Windows users) in the project directory.
- Run `go get` to download necessary dependencies.
- Run `go build` to build an executable file.
- Place the compiled executable anywhere on your computer.



You can also add the executable to path, which lets you use the executable from any directory on your computer:

- Add the path to that file in PATH environment
  variable. [Windows tutorial here.](https://stackoverflow.com/questions/44272416/how-to-add-a-folder-to-path-environment-variable-in-windows-10-with-screensho)
- Restart your command-line interface.

## How to Use

You can access the help command using `--help` or `-h`.

### Configuration file generation

Configuration file generation is fairly straightforward. The following command will create a configuration file in your
home directory.

```console
goub generate_config --client_id YOUR_CLIENT_ID --client_secret YOUR_CLIENT_SECRET
```

Obviously, replace `YOUR_CLIENT_ID` and `YOUR_CLIENT_SECRET` with your **osu! API v2** credentials.

### Downloading beatmaps

The following downloads all Ryuusei Aika's ranked and favorite beatmapsets.

```console
goub download -u 7777875 --ranked --favorite
```

You can also specify output path. The following command downloads all Ryuusei Aika's ranked and favorite beatmapsets to
my songs folder.

```console
goub download -u 7777875 -r -f -o C:\Users\Kris\AppData\Local\osu!\Songs
```

`goub --help` for more.

### Recursive Favorites

[**What is recursion?**](https://en.wikipedia.org/wiki/Recursion_(computer_science))

Recursively downloads your favorite beatmapsets and the beatmapset's author's favorite beatmapsets and so on. Usually takes a **VERY** long time without specifying the recursion depth limit.

The usage is straight-forward. The following command downloads Ryuusei Aika's favorite beatmaps etc.

```console
goub recursive_favorites -u 7777875
```

You **CAN** and it **IS** recommended to specify recursion depth limit using `-d` or `--depth` argument. For example:

```console
goub recursive_favorites -u 7777875 -d 3
```

## Todo

- ~~Ability to download top plays (im lazy)~~
- ~~Recursively download user's favorites and the favorite beatmapset's creator's favorite beatmapsets and so on.~~