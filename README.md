# Antibuild

A fast and customizable static site builder for the modern web. More details on https://build.antipy.com

## Usage

### Install

You have two installation options. You can either download a precompiled binary from the https://build.antipy.com/install (install instructions there) or install the latest beta version using go, if you have it installed. If you want to install go visit [the Golang website](https://golang.org/doc/install).

If you want to install using go enter this command into your terminal.
`go get -u -v https://github.com/antipy/antibuild`

To test that the installation is working you can run.
`antibuild version`

### Get started

To start a new project with antibuild you can run this command
`antibuild new`

You will be asked what you want to name the project. This will be the name of the folder the project will be generated in.

You will now be asked what generator template you want to use. Currently only the 'basic' template is supported.

You can also choose what [modules](https://build.antipy.com/modules) you want to install. You can install and remove modules later aswell. Modules add functionality to antibuild like data file parsing and template manipualtion. You can learn more about modules [here](https://build.antipy.com/modules).

To run antibuild in your new project navigate into the directory with
`cd <project_name>`
then run
`antibuild develop`

To find more info about how everything works, go to https://build.antipy.com/get-started.

## Licence

This project is licenced under the MIT licence. More details can be found in the licence file.
