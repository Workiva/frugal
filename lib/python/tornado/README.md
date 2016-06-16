# Frugal Tornado

Frugal Tornado is a Python implementation of Frugal using the [official NATS
client](https://github.com/nats-io/python-nats) which uses Tornado 4.2+ behind the
scenes.  Currently Frugal Tornado only supports Python 2.7.

## Using

```bash
pip install frugal_tornado
```
or add frugal_tornado to requirements.txt

## Contributing
1. Clone the repo 
2. Make a virutalenv `mkvirtualenv frugal-tornado -a /path/to/frugal/lib/python`
3. Install dependecies `make deps`
4. Write code, tests & create a pull requests
    a. Bonus: Automatically run tests on fail save with `make sniffer`
