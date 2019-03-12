# Antibuild

A fast and customizable static site builder for the modern web. More details on https://build.antipy.com

## Usage

### Install

You have two installation options. You can either download a precompiled binary from the https://build.antipy.com/install (install instructions there) or install the latest beta version using go, if you have it installed. If you want to install go visit [the Golang website](https://golang.org/doc/install).

If you want to install using go enter this command into your terminal.

```bash
go get -u -v https://github.com/antipy/antibuild
go install https://github.com/antipy/antibuild
```

To test that the installation is working you can run.

```bash
antibuild version
```

### Get started

To start a new project with antibuild you can run this command

```bash
antibuild new
```

You will be asked few basic questions about the project.

1. What should the name of the project be? _The name of the folder the project should be initalized in._
2. Choose a starting template: _The template we should use as the project baseline._
3. Select any modules you want to pre install now (can also not choose any): _Basic modules you want to use. You can install and remove modules later aswell._

Modules add functionality to antibuild like data file parsing and template manipualtion. You can learn more about modules [here](https://build.antipy.com/modules).

To run antibuild in your new project navigate into the generated directory and run antibuild in development mode.

```bash
cd <project_name>
antibuild develop
```

To find more info about how everything works, go to https://build.antipy.com/get-started.

## Licence

This project is licenced under the MIT licence. More details can be found in the licence file.
