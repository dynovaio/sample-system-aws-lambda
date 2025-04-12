![Sample-Code](https://gitlab.com/softbutterfly/open-source/open-source-office/-/raw/master/assets/dynova/dynova-open-source--banner--sample-code.png)

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg)](CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/badge/License-BSD_3--Clause-blue.svg)](LICENSE.txt)
[![Jupyter Book Badge](https://jupyterbook.org/badge.svg)](https://dynovaio.github.io/sample-system-aws-functions)

# AWS Lambda Samples

This repository contains sample code for AWS Lambda functions. The samples are designed to help you understand how to insturment your AWS Lambda functions using New Relic and OpenTelemetry.

## Requirements

* sdkman ([↗][href:sdkman])
* goenv ([↗][href:goenv])
* nvm ([↗][href:nvm])
* docker ([↗][href:docker])
* docker-compose ([↗][href:docker-compose])
* AWS CLI ([↗][href:awscli])
* New Relic account ([↗][href:newrelic])
* Visual Studio Code ([VSCode ↗][href:vscode]) with the AWS Toolkit extension

## Directory structure:

```
.
├── scripts
├── sample-dotnet8@otel
└── sample-golang@otel
```

* `scripts`: Contains scripts to build and deploy the samples.
* `sample-dotnet8@otel`: Contains the sample code for the .NET 8 Lambda function.
* `sample-golang@otel`: Contains the sample code for the Golang Lambda function.

## Usage

To use the samples, follow these steps:

1. Clone the repository:

   ```bash
   git clone https://github.com/dynova.io/open-source/open-source-office.git
   ```

## Contributing

Sugestions and contributions are welcome!

> Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.

For more information, please refer to the [Code of Conduct ↗][href:code_of_conduct].

## License

This project is licensed under the terms of the [BSD-3-Clause
↗][href:license] license.


[href:sdkman]: https://sdkman.io/
[href:goenv]: https://github.com/go-nv/goenv.git
[href:nvm]: https://github.com/nvm-sh/nvm
[href:docker]: https://docs.docker.com/get-docker/
[href:docker-compose]: https://docs.docker.com/compose/install/
[href:awscli]: https://aws.amazon.com/es/cli/
[href:newrelic]: https://newrelic.com/signup/
[href:license]: LICENSE.txt
[href:code_of_conduct]: CODE_OF_CONDUCT.md
[href:vscode]: https://code.visualstudio.com/
