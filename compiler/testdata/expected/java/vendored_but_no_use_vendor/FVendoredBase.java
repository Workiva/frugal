/**
 * Autogenerated by Frugal Compiler (3.15.3)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *
 * @generated
 */

package vendor_namespace.java;

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
import java.util.Objects;
import java.nio.ByteBuffer;
import java.util.Arrays;
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
import com.workiva.frugal.provider.FServiceClient;
import com.workiva.frugal.provider.FServiceProvider;
import com.workiva.frugal.transport.FTransport;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.protocol.TMessageType;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import java.util.Arrays;
import java.util.concurrent.*;


public class FVendoredBase {

	private static final Logger logger = LoggerFactory.getLogger(FVendoredBase.class);

	public interface Iface extends InternalIface {}

	/** For internal use only. Contains only the methods defined directly by the service. */
	public interface InternalIface {

	}

	public static class Client implements Iface {

		private InternalIface proxy;

		public Client(FServiceProvider provider, ServiceMiddleware... middleware) {
			InternalIface client = new InternalClient(provider);
			List<ServiceMiddleware> combined = new ArrayList<ServiceMiddleware>(Arrays.asList(middleware));
			combined.addAll(provider.getMiddleware());
			middleware = combined.toArray(new ServiceMiddleware[0]);
			proxy = InvocationHandler.composeMiddleware(client, InternalIface.class, middleware);
		}

	}

	private static class InternalClient extends FServiceClient implements InternalIface {
		public InternalClient(FServiceProvider provider) {
			super(provider);
		}
	}

	public static class Processor extends FBaseProcessor implements FProcessor {

		private Iface handler;

		public Processor(Iface iface, ServiceMiddleware... middleware) {
			handler = InvocationHandler.composeMiddleware(iface, Iface.class, middleware);
		}

		protected java.util.Map<String, FProcessorFunction> getProcessMap() {
			java.util.Map<String, FProcessorFunction> processMap = new java.util.HashMap<>();
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

	}

}