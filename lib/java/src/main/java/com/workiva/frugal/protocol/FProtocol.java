package com.workiva.frugal.protocol;

import com.workiva.frugal.internal.Headers;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.*;

import java.nio.ByteBuffer;

/**
 * FProtocol is Frugal's equivalent of Thrift's TProtocol. It defines the
 * serialization protocol used for messages, such as JSON, binary, etc. FProtocol
 * actually extends TProtocol and adds support for serializing FContext. In
 * practice, FProtocol simply wraps a TProtocol and uses Thrift's built-in
 * serialization. FContext is encoded before the TProtocol serialization of the
 * message using a simple binary protocol. See the protocol documentation for more
 * details.
 */
public class FProtocol extends TProtocol {

    private TProtocol wrapped;

    protected FProtocol(TProtocol proto) {
        super(proto.getTransport());
        wrapped = proto;
    }

    /**
     * Writes the request headers set on the given FContext into the protocol.
     *
     * @param context context with headers to write
     * @throws TException an error occurred while writing the headers
     */
    public void writeRequestHeader(FContext context) throws TException {
        wrapped.getTransport().write(Headers.encode(context.getRequestHeaders()));
    }

    /**
     * Reads the request headers on the protocol into a returned FContext.
     *
     * @return FContext with read headers
     * @throws TException an error occurred while reading the headers
     */
    public FContext readRequestHeader() throws TException {
        FContext ctx = FContext.withRequestHeaders(Headers.read(wrapped.getTransport()));
        // Put op id in response headers
        ctx.setResponseOpId(Long.toString(ctx.getOpId()));
        return ctx;
    }

    /**
     * Writes the response headers set on the given FContext into the protocol.
     *
     * @param context context with headers to write
     * @throws TException an error occurred while writing the headers
     */
    public void writeResponseHeader(FContext context) throws TException {
        wrapped.getTransport().write(Headers.encode(context.getResponseHeaders()));
    }

    /**
     * Reads the response headers on the protocol into the given FContext.
     *
     * @param context context to read headers into
     * @throws TException an error occurred while reading the headers
     */
    public void readResponseHeader(FContext context) throws TException {
        context.forceAddResponseHeaders(Headers.read(wrapped.getTransport()));
    }

    @Override
    public void writeMessageBegin(TMessage tMessage) throws TException {
        wrapped.writeMessageBegin(tMessage);
    }

    @Override
    public void writeMessageEnd() throws TException {
        wrapped.writeMessageEnd();
    }

    @Override
    public void writeStructBegin(TStruct tStruct) throws TException {
        wrapped.writeStructBegin(tStruct);
    }

    @Override
    public void writeStructEnd() throws TException {
        wrapped.writeStructEnd();
    }

    @Override
    public void writeFieldBegin(TField tField) throws TException {
        wrapped.writeFieldBegin(tField);
    }

    @Override
    public void writeFieldEnd() throws TException {
        wrapped.writeFieldEnd();
    }

    @Override
    public void writeFieldStop() throws TException {
        wrapped.writeFieldStop();
    }

    @Override
    public void writeMapBegin(TMap tMap) throws TException {
        wrapped.writeMapBegin(tMap);
    }

    @Override
    public void writeMapEnd() throws TException {
        wrapped.writeMapEnd();
    }

    @Override
    public void writeListBegin(TList tList) throws TException {
        wrapped.writeListBegin(tList);
    }

    @Override
    public void writeListEnd() throws TException {
        wrapped.writeListEnd();
    }

    @Override
    public void writeSetBegin(TSet tSet) throws TException {
        wrapped.writeSetBegin(tSet);
    }

    @Override
    public void writeSetEnd() throws TException {
        wrapped.writeSetEnd();
    }

    @Override
    public void writeBool(boolean b) throws TException {
        wrapped.writeBool(b);
    }

    @Override
    public void writeByte(byte b) throws TException {
        wrapped.writeByte(b);
    }

    @Override
    public void writeI16(short i) throws TException {
        wrapped.writeI16(i);
    }

    @Override
    public void writeI32(int i) throws TException {
        wrapped.writeI32(i);
    }

    @Override
    public void writeI64(long l) throws TException {
        wrapped.writeI64(l);
    }

    @Override
    public void writeDouble(double v) throws TException {
        wrapped.writeDouble(v);
    }

    @Override
    public void writeString(String s) throws TException {
        wrapped.writeString(s);
    }

    @Override
    public void writeBinary(ByteBuffer byteBuffer) throws TException {
        wrapped.writeBinary(byteBuffer);
    }

    @Override
    public TMessage readMessageBegin() throws TException {
        return wrapped.readMessageBegin();
    }

    @Override
    public void readMessageEnd() throws TException {
        wrapped.readMessageEnd();
    }

    @Override
    public TStruct readStructBegin() throws TException {
        return wrapped.readStructBegin();
    }

    @Override
    public void readStructEnd() throws TException {
        wrapped.readStructEnd();
    }

    @Override
    public TField readFieldBegin() throws TException {
        return wrapped.readFieldBegin();
    }

    @Override
    public void readFieldEnd() throws TException {
        wrapped.readFieldEnd();
    }

    @Override
    public TMap readMapBegin() throws TException {
        return wrapped.readMapBegin();
    }

    @Override
    public void readMapEnd() throws TException {
        wrapped.readMapEnd();
    }

    @Override
    public TList readListBegin() throws TException {
        return wrapped.readListBegin();
    }

    @Override
    public void readListEnd() throws TException {
        wrapped.readListEnd();
    }

    @Override
    public TSet readSetBegin() throws TException {
        return wrapped.readSetBegin();
    }

    @Override
    public void readSetEnd() throws TException {
        wrapped.readSetEnd();
    }

    @Override
    public boolean readBool() throws TException {
        return wrapped.readBool();
    }

    @Override
    public byte readByte() throws TException {
        return wrapped.readByte();
    }

    @Override
    public short readI16() throws TException {
        return wrapped.readI16();
    }

    @Override
    public int readI32() throws TException {
        return wrapped.readI32();
    }

    @Override
    public long readI64() throws TException {
        return wrapped.readI64();
    }

    @Override
    public double readDouble() throws TException {
        return wrapped.readDouble();
    }

    @Override
    public String readString() throws TException {
        return wrapped.readString();
    }

    @Override
    public ByteBuffer readBinary() throws TException {
        return wrapped.readBinary();
    }
}
