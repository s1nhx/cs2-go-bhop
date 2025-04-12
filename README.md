# cs2-go-bhop

> :warning: **Disclaimer:** I do **not** recommend trying to run cs2-go-bhop:
>
> - on your main Steam account;
> - without `-insecure` launch option;
> - with any external AC (such as FACEIT) turned on and loaded into game.
>
> **This project was made exclusively for educational purposes**. Your account **will** get banned if you try to boot up the game with cs2-go-bhop turned on.

## What is **cs2-go-bhop**?
<small>This explained more thoroughly [here](https://github.com/s1nhx/cs2-go-bhop-research).</small>\
**cs2-go-bhop** is a simple bunnyhop cheat for CS2 fully written in **Go**. Mainly, this project started because I wanted to try new programming language and, supposedly, should've worked only on x86 architecture, but was extended to build on x64 too.

## Workflow explanatory
Creation and workflow of this project explained [here](https://github.com/s1nhx/cs2-go-bhop-research).

## Usage
### Building a project
1. Clone this repository: `git clone https://github.com/s1nhx/cs2-go-bhop.git`
2. Build using `go build`: `cd cs2-go-bhop && go build`
3. Run builded file: `.\cs2-go-bhop.exe`

### Dependencies
Project uses only Go built-in functions, such as `syscall`, `encoding/binary`, etc. No other dependencies are required to install.
