/**
 * Autogenerated by Frugal Compiler (1.5.1)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *
 * @generated
 */

package base;

import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.exception.FTimeoutException;
import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.processor.FBaseProcessor;
import com.workiva.frugal.processor.FProcessor;
import com.workiva.frugal.processor.FProcessorFunction;
import com.workiva.frugal.protocol.*;
import com.workiva.frugal.transport.FTransport;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.protocol.TMessage;
import org.apache.thrift.protocol.TMessageType;
import org.apache.thrift.transport.TTransport;

import javax.annotation.Generated;
import java.util.concurrent.*;


@Generated(value = "Autogenerated by Frugal Compiler (1.5.1)", date = "2016-6-7")
public class FBaseFoo {

	public interface Iface {

		public void basePing(FContext ctx) throws TException;

	}

	public static class Client implements Iface {

		protected final Object writeLock = new Object();
		private Iface proxy;

		public Client(FTransport transport, FProtocolFactory protocolFactory, ServiceMiddleware... middleware) {
			Iface client = new InternalClient(transport, protocolFactory, writeLock);
			proxy = InvocationHandler.composeMiddleware(client, Iface.class, middleware);
		}

		public void basePing(FContext ctx) throws TException {
			proxy.basePing(ctx);
		}

	}

	private static class InternalClient implements Iface {

		private FTransport transport;
		private FProtocolFactory protocolFactory;
		private FProtocol inputProtocol;
		private FProtocol outputProtocol;
		private final Object writeLock;

		public InternalClient(FTransport transport, FProtocolFactory protocolFactory, Object writeLock) {
			this.transport = transport;
			this.transport.setRegistry(new FClientRegistry());
			this.protocolFactory = protocolFactory;
			this.inputProtocol = this.protocolFactory.getProtocol(this.transport);
			this.outputProtocol = this.protocolFactory.getProtocol(this.transport);
			this.writeLock = writeLock;
		}

		public void basePing(FContext ctx) throws TException {
			FProtocol oprot = this.outputProtocol;
			BlockingQueue<Object> result = new ArrayBlockingQueue<>(1);
			this.transport.register(ctx, recvBasePingHandler(ctx, result));
			try {
				synchronized (writeLock) {
					oprot.writeRequestHeader(ctx);
					oprot.writeMessageBegin(new TMessage("basePing", TMessageType.CALL, 0));
					BaseFoo.basePing_args args = new BaseFoo.basePing_args();
					args.write(oprot);
					oprot.writeMessageEnd();
					oprot.getTransport().flush();
				}

				Object res = null;
				try {
					res = result.poll(ctx.getTimeout(), TimeUnit.MILLISECONDS);
				} catch (InterruptedException e) {
					throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "basePing interrupted: " + e.getMessage());
				}
				if (res == null) {
					throw new FTimeoutException("basePing timed out");
				}
				if (res instanceof TException) {
					throw (TException) res;
				}
				BaseFoo.basePing_result r = (BaseFoo.basePing_result) res;
			} finally {
				this.transport.unregister(ctx);
			}
		}

		private FAsyncCallback recvBasePingHandler(final FContext ctx, final BlockingQueue<Object> result) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = InternalClient.this.protocolFactory.getProtocol(tr);
					try {
						iprot.readResponseHeader(ctx);
						TMessage message = iprot.readMessageBegin();
						if (!message.name.equals("basePing")) {
							throw new TApplicationException(TApplicationException.WRONG_METHOD_NAME, "basePing failed: wrong method name");
						}
						if (message.type == TMessageType.EXCEPTION) {
							TApplicationException e = TApplicationException.read(iprot);
							iprot.readMessageEnd();
							if (e.getType() == FTransport.RESPONSE_TOO_LARGE) {
								FMessageSizeException ex = new FMessageSizeException(FTransport.RESPONSE_TOO_LARGE, "response too large for transport");
								try {
									result.put(ex);
									return;
								} catch (InterruptedException ie) {
									throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "basePing interrupted: " + ie.getMessage());
								}
							}
							try {
								result.put(e);
							} finally {
								throw e;
							}
						}
						if (message.type != TMessageType.REPLY) {
							throw new TApplicationException(TApplicationException.INVALID_MESSAGE_TYPE, "basePing failed: invalid message type");
						}
						BaseFoo.basePing_result res = new BaseFoo.basePing_result();
						res.read(iprot);
						iprot.readMessageEnd();
						try {
							result.put(res);
						} catch (InterruptedException e) {
							throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "basePing interrupted: " + e.getMessage());
						}
					} catch (TException e) {
						try {
							result.put(e);
						} finally {
							throw e;
						}
					}
				}
			};
		}

	}

	public static class Processor extends FBaseProcessor implements FProcessor {

		public Processor(Iface iface, ServiceMiddleware... middleware) {
			super(getProcessMap(iface, new java.util.HashMap<String, FProcessorFunction>(), middleware));
		}

		protected Processor(Iface iface, java.util.Map<String, FProcessorFunction> processMap, ServiceMiddleware[] middleware) {
			super(getProcessMap(iface, processMap, middleware));
		}

		private static java.util.Map<String, FProcessorFunction> getProcessMap(Iface handler, java.util.Map<String, FProcessorFunction> processMap, ServiceMiddleware[] middleware) {
			handler = InvocationHandler.composeMiddleware(handler, Iface.class, middleware);
			processMap.put("basePing", new BasePing(handler));
			return processMap;
		}

		private static class BasePing implements FProcessorFunction {

			private Iface handler;

			public BasePing(Iface handler) {
				this.handler = handler;
			}

			public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {
				BaseFoo.basePing_args args = new BaseFoo.basePing_args();
				try {
					args.read(iprot);
				} catch (TException e) {
					iprot.readMessageEnd();
					synchronized (WRITE_LOCK) {
						writeApplicationException(ctx, oprot, TApplicationException.PROTOCOL_ERROR, "basePing", e.getMessage());
					}
					throw e;
				}

				iprot.readMessageEnd();
				BaseFoo.basePing_result result = new BaseFoo.basePing_result();
				try {
					this.handler.basePing(ctx);
				} catch (TException e) {
					synchronized (WRITE_LOCK) {
						writeApplicationException(ctx, oprot, TApplicationException.INTERNAL_ERROR, "basePing", "Internal error processing basePing: " + e.getMessage());
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
					} catch (TException e) {
						if (e instanceof FMessageSizeException) {
							writeApplicationException(ctx, oprot, FTransport.RESPONSE_TOO_LARGE, "basePing", "response too large: " + e.getMessage());
						} else {
							throw e;
						}
					}
				}
			}
		}

		private static void writeApplicationException(FContext ctx, FProtocol oprot, int type, String method, String message) throws TException {
			TApplicationException x = new TApplicationException(type, message);
			oprot.writeResponseHeader(ctx);
			oprot.writeMessageBegin(new TMessage(method, TMessageType.EXCEPTION, 0));
			x.write(oprot);
			oprot.writeMessageEnd();
			oprot.getTransport().flush();
		}

	}

}