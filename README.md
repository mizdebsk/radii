[![build status](https://img.shields.io/github/actions/workflow/status/mizdebsk/radii/ci.yml?branch=main)](https://github.com/mizdebsk/radii/actions/workflows/ci.yml?query=branch%3Amain)
[![License](https://img.shields.io/github/license/mizdebsk/radii.svg?label=License)](https://www.gnu.org/licenses/gpl-3.0-standalone.html)


radii
=====

Hardware driver manager for Linux

The radii tool is a command-line utility that provides a consistent
interface for installing and maintaining third-party AI-accelerator
and GPU driver stacks on Linux systems.  It detects compatible
accelerator hardware, enables the appropriate package repositories,
and installs the required kernel and user-mode components from
packages provided by the Linux distribution.  The tool manages
multi-component driver stacks end-to-end and integrates with the
standard system package manager workflow so that installed drivers are
updated through normal system package management.

The name "radii" comes from RDI (RHEL Drivers Installation), the core
purpose of the tool.  It echoes the acronym while also being a real
word.  Radii is the plural of the Latin radius, which originally meant
a ray, rod, or spoke.  The plural form reflects the many pieces that
make up a driver stack.  In anatomy, the radii are the bones that let
your hands move, serving as a simple metaphor for enabling hardware to
work.

Licensed under GPL v3 or later.
