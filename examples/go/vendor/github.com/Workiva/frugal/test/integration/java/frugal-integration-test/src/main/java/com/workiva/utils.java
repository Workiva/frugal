package com.workiva;

import org.apache.thrift.protocol.TBinaryProtocol;
import org.apache.thrift.protocol.TCompactProtocol;
import org.apache.thrift.protocol.TJSONProtocol;
import org.apache.thrift.protocol.TProtocolFactory;

import java.util.ArrayList;
import java.util.List;


public class utils {

    public static String PREAMBLE_HEADER = "preamble";
    public static String RAMBLE_HEADER = "ramble";

    public static TProtocolFactory whichProtocolFactory (String protocol_type) throws Exception {
        List<String> validProtocols = new ArrayList<>();
        validProtocols.add("binary");
        validProtocols.add("compact");
        validProtocols.add("json");

        if (!validProtocols.contains(protocol_type)) {
            throw new Exception("Unknown protocol type! " + protocol_type);
        }

        TProtocolFactory protocolFactory;
        switch (protocol_type) {
            case "json":
                protocolFactory = new TJSONProtocol.Factory();
                break;
            case "compact":
                protocolFactory = new TCompactProtocol.Factory();
                break;
            case "binary":
                protocolFactory = new TBinaryProtocol.Factory();
                break;
            default:
                throw new Exception("Unknown protocol type encountered: " + protocol_type);
        }
        return protocolFactory;
    }

}
