package thrift;

import org.apache.thrift.TException;
import org.apache.thrift.server.TServer;
import org.apache.thrift.server.TSimpleServer;
import org.apache.thrift.transport.TServerSocket;
import org.apache.thrift.transport.TServerTransport;

public class ServerMain {

    public static TestThriftService.Iface handler;

    public static TestThriftService.Processor processor;

    public static void main(String[] args) {
        try {
            handler = new TestThriftServiceHandler();
            processor = new TestThriftService.Processor(handler);

            Runnable simple = new Runnable() {
                public void run() {
                    simple(processor);
                }
            };

            Thread thread = new Thread(simple);
            thread.start();
        } catch (Exception x) {
            x.printStackTrace();
            throw x;
        }
    }

    public static void simple(TestThriftService.Processor processor) {
        try {
            TServerTransport serverTransport = new TServerSocket(9090);
            TServer server = new TSimpleServer(new TServer.Args(serverTransport).processor(processor));

            System.out.println("Starting the simple server...");
            server.serve();
            System.out.println("please no");
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
