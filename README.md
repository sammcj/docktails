# Docktails

_"It's Docktails Time!"_

Annoyed when docker logs -f exits when a container restarts? Want to see the logs of multiple containers at once? Docktails might be for you!


## Installation

Download the latest release from the releases page, or build it yourself with `make build`.

## Usage

```shell
docktails [CONTAINER] [CONTAINER]...
```

If no containers are specified, docktails will show the logs of all running containers for you to choose from.

```log
$ docktails textgen whalewall

Tailing logs for containers: textgen, whalewall
-------------
| whalewall |
-------------
[info] [2023-08-31 11:18:15] X{"level":"info","time":"2023-08-30T05:25:53.836460879Z","msg":"applied landlock rules"}
[info] [2023-08-31 11:18:15] o{"level":"info","time":"2023-08-30T05:25:53.836756758Z","msg":"applied seccomp filters","syscalls.allowed":48}

-----------
| textgen |
-----------
[info] [2023-08-31 11:18:15] Running on local URL:  http://127.0.0.1:7860
[info] [2023-08-31 11:18:15]
[info] [2023-08-31 11:18:15] To create a public link
[info] [2023-08-31 11:18:15] , set `share=True` in `launch()`.
Running on local URL:  http://127.0.0.1:7860
```

```log
$ docktails

Select containers (comma-separated, e.g., 1,2,3):
[1] textgen
[2] langflow
[3] whalewall
Enter the numbers of the containers to tail logs (comma-separated):
```

## Contributing

Pull requests are welcome! For major changes, please open an issue first to discuss what you would like to change.

## License and Acknowledgements

- [MIT](LICENSE)
- Copyright (c) 2023 Sam McLeod