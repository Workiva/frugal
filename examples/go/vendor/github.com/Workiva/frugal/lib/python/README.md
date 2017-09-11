# Frugal Python

## Using

```bash
pip install frugal==1.8.0
```
or to use with the [Tornado nats-client](https://github.com/nats-io/python-nats)
```bash
pip install "frugal[tornado]"==1.8.0
```
or preferably add one of the following to requirements.txt
```bash
frugal==1.8.0
# or for tornado support
frugal[tornado]==1.8.0
```
## Contributing
1. Make a virutalenv `mkvirtualenv frugal -a /path/to/frugal/lib/python`
2. Install dependecies `make deps`
3. Write code, tests & create a pull requests
    * Automatically run tests on fail save with `make sniffer`
