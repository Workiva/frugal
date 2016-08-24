### Frugal Python Tornado Example
=====

1. Highly recommend making a virtualenv with virtualenvwrapper:
    ```
    mkvirtualenv frugal-example -a $PWD
    ```

1. Install deps:
    ```
    pip install -Ur requirements.txt
    ```

1. Install and/or Start up NATS:
    ```bash
    brew install gnatsd
    # Follow instructions to run as daemon or...
    gnatsd -DV
    ```

1. Start up server:
    ```
    python run_server.py
    ```

1. Run client:
    ```
    python main.py
    ```
