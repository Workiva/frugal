package thrift;

import org.apache.thrift.TException;

public class TestThriftServiceHandler implements TestThriftService.Iface {
    @Override
    public String testString(String s) throws TException {
        System.out.println(s);
        return s;
    }
}
