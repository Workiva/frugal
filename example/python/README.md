### Frugal Python Example
=====

1. Highly recommend making a virtualenv with virtualenvwrapper:
    ```
    mkvirtualenv frugal-example -a $PWD
    ```

2. Install deps:
    ```
    pip install -Ur requirements.txt
    ```

3. Install and/or Start Up NATS:
    ```bash
    brew install gnatsd
    # Follow instructions to run as daemon or...
    gnatsd -DV
    ```

4. Profit:
    ```
    python main.py
    ```
