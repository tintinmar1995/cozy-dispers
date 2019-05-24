Cozy DISPERS
==========

[![GoDoc](https://godoc.org/github.com/cozy/cozy-stack?status.svg)](https://godoc.org/github.com/cozy/cozy-stack)
[![Build Status](https://travis-ci.org/cozy/cozy-stack.svg?branch=master)](https://travis-ci.org/cozy/cozy-stack)
[![Go Report Card](https://goreportcard.com/badge/github.com/cozy/cozy-stack)](https://goreportcard.com/report/github.com/cozy/cozy-stack)


## What is Cozy?

![Cozy Logo](https://cdn.rawgit.com/cozy/cozy-guidelines/master/templates/cozy_logo_small.svg)

[Cozy](https://cozy.io) is a platform that brings all your web services in the
same private space. With it, your web apps and your devices can share data
easily, providing you with a new experience. You can install Cozy on your own
hardware where no one profiles you.


## What is Cozy-DISPERS

It is the core server of the Cozy-DISPERs platform.

Cozy-DISPERS is in charge of serving a query over a list of stacks. This process is a privacy-by-design algorithm.

It provides its services through a REST API that allows to:

 - subscribe, or unsubscribe to queries
 - launch a query or even a machine learning algorithm


Five actors are available:

 - Conductor
 - Concept Indexor
 - Target Finder
 - Target
 - Data Aggregator


Feel free to [open an issue](https://github.com/cozy/cozy-dispers/issues/new)
for questions and suggestions.


## Installing `cozy-dispers`

It is quite like installing a cozy-stack

You can follow the [Install guide](docs/INSTALL.md) and the [configuration
documentation](docs/config.md).


## How to contribute?

We are eager for contributions and very happy when we receive them! It can
code, of course, but it can also take other forms. The workflow is explained
in [the contributing guide](docs/CONTRIBUTING.md).


## Community

You can reach the Cozy Community by:

* Chatting with us on IRC #cozycloud on irc.freenode.net
* Posting on our [Forum](https://forum.cozy.io)
* Posting issues on the [Github repos](https://github.com/cozy/)
* Mentioning us on [Twitter](https://twitter.com/cozycloud)


## License

Cozy is developed by Cozy Cloud and distributed under the AGPL v3 license.
