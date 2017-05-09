/**
 * Autogenerated by Frugal Compiler (2.4.0)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *
 * @generated
 */

package actual_base.java;

import org.apache.thrift.scheme.IScheme;
import org.apache.thrift.scheme.SchemeFactory;
import org.apache.thrift.scheme.StandardScheme;

import org.apache.thrift.scheme.TupleScheme;
import org.apache.thrift.protocol.TTupleProtocol;
import org.apache.thrift.protocol.TProtocolException;
import org.apache.thrift.EncodingUtils;
import org.apache.thrift.TException;
import org.apache.thrift.async.AsyncMethodCallback;
import org.apache.thrift.server.AbstractNonblockingServer.*;
import java.util.List;
import java.util.ArrayList;
import java.util.Map;
import java.util.HashMap;
import java.util.EnumMap;
import java.util.Set;
import java.util.HashSet;
import java.util.EnumSet;
import java.util.Collections;
import java.util.BitSet;
import java.nio.ByteBuffer;
import java.util.Arrays;
import javax.annotation.Generated;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TApplicationExceptionType;
import com.workiva.frugal.exception.TTransportExceptionType;
import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.processor.FBaseProcessor;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.processor.FProcessorFunction;
import com.workiva.frugal.protocol.*;
import com.workiva.frugal.provider.FServiceProvider;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.protocol.TMessageType;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import javax.annotation.Generated;
import java.util.Arrays;
import java.util.concurrent.*;


@Generated(value = "Autogenerated by Frugal Compiler (2.4.0)", date = "2015-11-24")
public class FBaseFoo {

	private static final Logger logger = LoggerFactory.getLogger(FBaseFoo.class);

	public interface Iface {

		public void basePing(FContext ctx) throws TException;

	}

	public static class Client implements Iface {

		private Iface proxy;

		public Client(FServiceProvider provider, ServiceMiddleware... middleware) {
			Iface client = new InternalClient(provider);
			List<ServiceMiddleware> combined = Arrays.asList(middleware);
			combined.addAll(provider.getMiddleware());
			middleware = combined.toArray(new ServiceMiddleware[0]);
			proxy = InvocationHandler.composeMiddleware(client, Iface.class, middleware);
		}

		public void basePing(FContext ctx) throws TException {
			proxy.basePing(ctx);
		}

	}

	private static class InternalClient implements Iface {

		private FTransport transport;
		private FProtocolFactory protocolFactory;
		public InternalClient(FServiceProvider provider) {
			this.transport = provider.getTransport();
			this.protocolFactory = provider.getProtocolFactory();
		}

		public void basePing(FContext ctx) throws TException {
			TMemoryOutputBuffer memoryBuffer = new TMemoryOutputBuffer(this.transport.getRequestSizeLimit());
			FProtocol oprot = this.protocolFactory.getProtocol(memoryBuffer);
			oprot.writeRequestHeader(ctx);
			oprot.writeMessageBegin(new TMessage("basePing", TMessageType.CALL, 0));
			basePing_args args = new basePing_args();
			args.write(oprot);
			oprot.writeMessageEnd();
			TTransport response = this.transport.request(ctx, memoryBuffer.getWriteBytes());

			FProtocol iprot = this.protocolFactory.getProtocol(response);
			iprot.readResponseHeader(ctx);
			TMessage message = iprot.readMessageBegin();
			if (!message.name.equals("basePing")) {
				throw new TApplicationException(TApplicationExceptionType.WRONG_METHOD_NAME, "basePing failed: wrong method name");
			}
			if (message.type == TMessageType.EXCEPTION) {
				TApplicationException e = TApplicationException.read(iprot);
				iprot.readMessageEnd();
				TException returnedException = e;
				if (e.getType() == TApplicationExceptionType.RESPONSE_TOO_LARGE) {
					returnedException = new TTransportException(TTransportExceptionType.RESPONSE_TOO_LARGE, e.getMessage());
				}
				throw returnedException;
			}
			if (message.type != TMessageType.REPLY) {
				throw new TApplicationException(TApplicationExceptionType.INVALID_MESSAGE_TYPE, "basePing failed: invalid message type");
			}
			basePing_result res = new basePing_result();
			res.read(iprot);
			iprot.readMessageEnd();
		}
	}

	public static class Processor extends FBaseProcessor implements FProcessor {

		private Iface handler;

		public Processor(Iface iface, ServiceMiddleware... middleware) {
			handler = InvocationHandler.composeMiddleware(iface, Iface.class, middleware);
		}

		protected java.util.Map<String, FProcessorFunction> getProcessMap() {
			java.util.Map<String, FProcessorFunction> processMap = new java.util.HashMap<>();
			processMap.put("basePing", new BasePing());
			return processMap;
		}

		protected java.util.Map<String, java.util.Map<String, String>> getAnnotationsMap() {
			java.util.Map<String, java.util.Map<String, String>> annotationsMap = new java.util.HashMap<>();
			return annotationsMap;
		}

		@Override
		public void addMiddleware(ServiceMiddleware middleware) {
			handler = InvocationHandler.composeMiddleware(handler, Iface.class, new ServiceMiddleware[]{middleware});
		}

		private class BasePing implements FProcessorFunction {

			public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {
				basePing_args args = new basePing_args();
				try {
					args.read(iprot);
				} catch (TException e) {
					iprot.readMessageEnd();
					synchronized (WRITE_LOCK) {
						e = writeApplicationException(ctx, oprot, TApplicationExceptionType.PROTOCOL_ERROR, "basePing", e.getMessage());
					}
					throw e;
				}

				iprot.readMessageEnd();
				basePing_result result = new basePing_result();
				try {
					handler.basePing(ctx);
				} catch (TApplicationException e) {
					oprot.writeResponseHeader(ctx);
					oprot.writeMessageBegin(new TMessage("basePing", TMessageType.EXCEPTION, 0));
					e.write(oprot);
					oprot.writeMessageEnd();
					oprot.getTransport().flush();
					return;
				} catch (TException e) {
					synchronized (WRITE_LOCK) {
						e = writeApplicationException(ctx, oprot, TApplicationExceptionType.INTERNAL_ERROR, "basePing", "Internal error processing basePing: " + e.getMessage());
					}
					throw e;
				}
				synchronized (WRITE_LOCK) {
					try {
						oprot.writeResponseHeader(ctx);
						oprot.writeMessageBegin(new TMessage("basePing", TMessageType.REPLY, 0));
						result.write(oprot);
						oprot.writeMessageEnd();
						oprot.getTransport().flush();
					} catch (TTransportException e) {
						if (e.getType() == TTransportExceptionType.REQUEST_TOO_LARGE) {
							writeApplicationException(ctx, oprot, TApplicationExceptionType.RESPONSE_TOO_LARGE, "basePing", "response too large: " + e.getMessage());
						} else {
							throw e;
						}
					}
				}
			}
		}

	}

public static class basePing_args implements org.apache.thrift.TBase<basePing_args, basePing_args._Fields>, java.io.Serializable, Cloneable, Comparable<basePing_args> {
	private static final org.apache.thrift.protocol.TStruct STRUCT_DESC = new org.apache.thrift.protocol.TStruct("basePing_args");


	private static final Map<Class<? extends IScheme>, SchemeFactory> schemes = new HashMap<Class<? extends IScheme>, SchemeFactory>();
	static {
		schemes.put(StandardScheme.class, new basePing_argsStandardSchemeFactory());
		schemes.put(TupleScheme.class, new basePing_argsTupleSchemeFactory());
	}

	/** The set of fields this struct contains, along with convenience methods for finding and manipulating them. */
	public enum _Fields implements org.apache.thrift.TFieldIdEnum {
;

		private static final Map<String, _Fields> byName = new HashMap<String, _Fields>();

		static {
			for (_Fields field : EnumSet.allOf(_Fields.class)) {
				byName.put(field.getFieldName(), field);
			}
		}

		/**
		 * Find the _Fields constant that matches fieldId, or null if its not found.
		 */
		public static _Fields findByThriftId(int fieldId) {
			switch(fieldId) {
				default:
					return null;
			}
		}

		/**
		 * Find the _Fields constant that matches fieldId, throwing an exception
		 * if it is not found.
		 */
		public static _Fields findByThriftIdOrThrow(int fieldId) {
			_Fields fields = findByThriftId(fieldId);
			if (fields == null) throw new IllegalArgumentException("Field " + fieldId + " doesn't exist!");
			return fields;
		}

		/**
		 * Find the _Fields constant that matches name, or null if its not found.
		 */
		public static _Fields findByName(String name) {
			return byName.get(name);
		}

		private final short _thriftId;
		private final String _fieldName;

		_Fields(short thriftId, String fieldName) {
			_thriftId = thriftId;
			_fieldName = fieldName;
		}

		public short getThriftFieldId() {
			return _thriftId;
		}

		public String getFieldName() {
			return _fieldName;
		}
	}

	// isset id assignments
	public basePing_args() {
	}

	/**
	 * Performs a deep copy on <i>other</i>.
	 */
	public basePing_args(basePing_args other) {
	}

	public basePing_args deepCopy() {
		return new basePing_args(this);
	}

	@Override
	public void clear() {
	}

	public void setFieldValue(_Fields field, Object value) {
		switch (field) {
		}
	}

	public Object getFieldValue(_Fields field) {
		switch (field) {
		}
		throw new IllegalStateException();
	}

	/** Returns true if field corresponding to fieldID is set (has been assigned a value) and false otherwise */
	public boolean isSet(_Fields field) {
		if (field == null) {
			throw new IllegalArgumentException();
		}

		switch (field) {
		}
		throw new IllegalStateException();
	}

	@Override
	public boolean equals(Object that) {
		if (that == null)
			return false;
		if (that instanceof basePing_args)
			return this.equals((basePing_args)that);
		return false;
	}

	public boolean equals(basePing_args that) {
		if (that == null)
			return false;

		return true;
	}

	@Override
	public int hashCode() {
		List<Object> list = new ArrayList<Object>();

		return list.hashCode();
	}

	@Override
	public int compareTo(basePing_args other) {
		if (!getClass().equals(other.getClass())) {
			return getClass().getName().compareTo(other.getClass().getName());
		}

		int lastComparison = 0;

		return 0;
	}

	public _Fields fieldForId(int fieldId) {
		return _Fields.findByThriftId(fieldId);
	}

	public void read(org.apache.thrift.protocol.TProtocol iprot) throws org.apache.thrift.TException {
		schemes.get(iprot.getScheme()).getScheme().read(iprot, this);
	}

	public void write(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {
		schemes.get(oprot.getScheme()).getScheme().write(oprot, this);
	}

	@Override
	public String toString() {
		StringBuilder sb = new StringBuilder("basePing_args(");
		boolean first = true;

		sb.append(")");
		return sb.toString();
	}

	public void validate() throws org.apache.thrift.TException {
		// check for required fields
		// check for sub-struct validity
	}

	private void writeObject(java.io.ObjectOutputStream out) throws java.io.IOException {
		try {
			write(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(out)));
		} catch (org.apache.thrift.TException te) {
			throw new java.io.IOException(te);
		}
	}

	private void readObject(java.io.ObjectInputStream in) throws java.io.IOException, ClassNotFoundException {
		try {
			// it doesn't seem like you should have to do this, but java serialization is wacky, and doesn't call the default constructor.
			read(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(in)));
		} catch (org.apache.thrift.TException te) {
			throw new java.io.IOException(te);
		}
	}

	private static class basePing_argsStandardSchemeFactory implements SchemeFactory {
		public basePing_argsStandardScheme getScheme() {
			return new basePing_argsStandardScheme();
		}
	}

	private static class basePing_argsStandardScheme extends StandardScheme<basePing_args> {

		public void read(org.apache.thrift.protocol.TProtocol iprot, basePing_args struct) throws org.apache.thrift.TException {
			org.apache.thrift.protocol.TField schemeField;
			iprot.readStructBegin();
			while (true) {
				schemeField = iprot.readFieldBegin();
				if (schemeField.type == org.apache.thrift.protocol.TType.STOP) {
					break;
				}
				switch (schemeField.id) {
					default:
						org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);
				}
				iprot.readFieldEnd();
			}
			iprot.readStructEnd();

			// check for required fields of primitive type, which can't be checked in the validate method
			struct.validate();
		}

		public void write(org.apache.thrift.protocol.TProtocol oprot, basePing_args struct) throws org.apache.thrift.TException {
			struct.validate();

			oprot.writeStructBegin(STRUCT_DESC);
			oprot.writeFieldStop();
			oprot.writeStructEnd();
		}

	}

	private static class basePing_argsTupleSchemeFactory implements SchemeFactory {
		public basePing_argsTupleScheme getScheme() {
			return new basePing_argsTupleScheme();
		}
	}

	private static class basePing_argsTupleScheme extends TupleScheme<basePing_args> {

		@Override
		public void write(org.apache.thrift.protocol.TProtocol prot, basePing_args struct) throws org.apache.thrift.TException {
			TTupleProtocol oprot = (TTupleProtocol) prot;
		}

		@Override
		public void read(org.apache.thrift.protocol.TProtocol prot, basePing_args struct) throws org.apache.thrift.TException {
			TTupleProtocol iprot = (TTupleProtocol) prot;
		}

	}

}
public static class basePing_result implements org.apache.thrift.TBase<basePing_result, basePing_result._Fields>, java.io.Serializable, Cloneable, Comparable<basePing_result> {
	private static final org.apache.thrift.protocol.TStruct STRUCT_DESC = new org.apache.thrift.protocol.TStruct("basePing_result");


	private static final Map<Class<? extends IScheme>, SchemeFactory> schemes = new HashMap<Class<? extends IScheme>, SchemeFactory>();
	static {
		schemes.put(StandardScheme.class, new basePing_resultStandardSchemeFactory());
		schemes.put(TupleScheme.class, new basePing_resultTupleSchemeFactory());
	}

	/** The set of fields this struct contains, along with convenience methods for finding and manipulating them. */
	public enum _Fields implements org.apache.thrift.TFieldIdEnum {
;

		private static final Map<String, _Fields> byName = new HashMap<String, _Fields>();

		static {
			for (_Fields field : EnumSet.allOf(_Fields.class)) {
				byName.put(field.getFieldName(), field);
			}
		}

		/**
		 * Find the _Fields constant that matches fieldId, or null if its not found.
		 */
		public static _Fields findByThriftId(int fieldId) {
			switch(fieldId) {
				default:
					return null;
			}
		}

		/**
		 * Find the _Fields constant that matches fieldId, throwing an exception
		 * if it is not found.
		 */
		public static _Fields findByThriftIdOrThrow(int fieldId) {
			_Fields fields = findByThriftId(fieldId);
			if (fields == null) throw new IllegalArgumentException("Field " + fieldId + " doesn't exist!");
			return fields;
		}

		/**
		 * Find the _Fields constant that matches name, or null if its not found.
		 */
		public static _Fields findByName(String name) {
			return byName.get(name);
		}

		private final short _thriftId;
		private final String _fieldName;

		_Fields(short thriftId, String fieldName) {
			_thriftId = thriftId;
			_fieldName = fieldName;
		}

		public short getThriftFieldId() {
			return _thriftId;
		}

		public String getFieldName() {
			return _fieldName;
		}
	}

	// isset id assignments
	public basePing_result() {
	}

	/**
	 * Performs a deep copy on <i>other</i>.
	 */
	public basePing_result(basePing_result other) {
	}

	public basePing_result deepCopy() {
		return new basePing_result(this);
	}

	@Override
	public void clear() {
	}

	public void setFieldValue(_Fields field, Object value) {
		switch (field) {
		}
	}

	public Object getFieldValue(_Fields field) {
		switch (field) {
		}
		throw new IllegalStateException();
	}

	/** Returns true if field corresponding to fieldID is set (has been assigned a value) and false otherwise */
	public boolean isSet(_Fields field) {
		if (field == null) {
			throw new IllegalArgumentException();
		}

		switch (field) {
		}
		throw new IllegalStateException();
	}

	@Override
	public boolean equals(Object that) {
		if (that == null)
			return false;
		if (that instanceof basePing_result)
			return this.equals((basePing_result)that);
		return false;
	}

	public boolean equals(basePing_result that) {
		if (that == null)
			return false;

		return true;
	}

	@Override
	public int hashCode() {
		List<Object> list = new ArrayList<Object>();

		return list.hashCode();
	}

	@Override
	public int compareTo(basePing_result other) {
		if (!getClass().equals(other.getClass())) {
			return getClass().getName().compareTo(other.getClass().getName());
		}

		int lastComparison = 0;

		return 0;
	}

	public _Fields fieldForId(int fieldId) {
		return _Fields.findByThriftId(fieldId);
	}

	public void read(org.apache.thrift.protocol.TProtocol iprot) throws org.apache.thrift.TException {
		schemes.get(iprot.getScheme()).getScheme().read(iprot, this);
	}

	public void write(org.apache.thrift.protocol.TProtocol oprot) throws org.apache.thrift.TException {
		schemes.get(oprot.getScheme()).getScheme().write(oprot, this);
	}

	@Override
	public String toString() {
		StringBuilder sb = new StringBuilder("basePing_result(");
		boolean first = true;

		sb.append(")");
		return sb.toString();
	}

	public void validate() throws org.apache.thrift.TException {
		// check for required fields
		// check for sub-struct validity
	}

	private void writeObject(java.io.ObjectOutputStream out) throws java.io.IOException {
		try {
			write(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(out)));
		} catch (org.apache.thrift.TException te) {
			throw new java.io.IOException(te);
		}
	}

	private void readObject(java.io.ObjectInputStream in) throws java.io.IOException, ClassNotFoundException {
		try {
			// it doesn't seem like you should have to do this, but java serialization is wacky, and doesn't call the default constructor.
			read(new org.apache.thrift.protocol.TCompactProtocol(new org.apache.thrift.transport.TIOStreamTransport(in)));
		} catch (org.apache.thrift.TException te) {
			throw new java.io.IOException(te);
		}
	}

	private static class basePing_resultStandardSchemeFactory implements SchemeFactory {
		public basePing_resultStandardScheme getScheme() {
			return new basePing_resultStandardScheme();
		}
	}

	private static class basePing_resultStandardScheme extends StandardScheme<basePing_result> {

		public void read(org.apache.thrift.protocol.TProtocol iprot, basePing_result struct) throws org.apache.thrift.TException {
			org.apache.thrift.protocol.TField schemeField;
			iprot.readStructBegin();
			while (true) {
				schemeField = iprot.readFieldBegin();
				if (schemeField.type == org.apache.thrift.protocol.TType.STOP) {
					break;
				}
				switch (schemeField.id) {
					default:
						org.apache.thrift.protocol.TProtocolUtil.skip(iprot, schemeField.type);
				}
				iprot.readFieldEnd();
			}
			iprot.readStructEnd();

			// check for required fields of primitive type, which can't be checked in the validate method
			struct.validate();
		}

		public void write(org.apache.thrift.protocol.TProtocol oprot, basePing_result struct) throws org.apache.thrift.TException {
			struct.validate();

			oprot.writeStructBegin(STRUCT_DESC);
			oprot.writeFieldStop();
			oprot.writeStructEnd();
		}

	}

	private static class basePing_resultTupleSchemeFactory implements SchemeFactory {
		public basePing_resultTupleScheme getScheme() {
			return new basePing_resultTupleScheme();
		}
	}

	private static class basePing_resultTupleScheme extends TupleScheme<basePing_result> {

		@Override
		public void write(org.apache.thrift.protocol.TProtocol prot, basePing_result struct) throws org.apache.thrift.TException {
			TTupleProtocol oprot = (TTupleProtocol) prot;
		}

		@Override
		public void read(org.apache.thrift.protocol.TProtocol prot, basePing_result struct) throws org.apache.thrift.TException {
			TTupleProtocol iprot = (TTupleProtocol) prot;
		}

	}

}
}