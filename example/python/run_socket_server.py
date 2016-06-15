from thrift.transport.TSocket import TServerSocket


def main():
    # server socket extends TTransportBase
    server_socket = TServerSocket()

    frugal_server = FSimpleServer(processor_factory,
                                  server_socket,
                                  transport_factory,
                                  protocol_factory)

    frugal_server.listen()
