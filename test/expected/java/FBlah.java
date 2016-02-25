/**
 * Autogenerated by Frugal Compiler (1.0.2)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *  @generated
 */

package foo;

import com.workiva.frugal.exception.FMessageSizeException;
import com.workiva.frugal.exception.FTimeoutException;
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
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.TimeUnit;


@Generated(value = "Autogenerated by Frugal Compiler (1.0.2)", date = "2015-11-24")
public class FBlah {

	public interface Iface {

		/**
		 * Use this to ping the server.
		 */
		public void ping(FContext ctx) throws TException;

		/**
		 * Use this to tell the sever how you feel.
		 */
		public long bleh(FContext ctx, Thing one, Stuff Two) throws TException, InvalidOperation;

}

	public static class Client implements Iface {

		private static final Object WRITE_LOCK = new Object();

		private FTransport transport;
		private FProtocolFactory protocolFactory;
		private FProtocol inputProtocol;
		private FProtocol outputProtocol;

		public Client(FTransport transport, FProtocolFactory protocolFactory) {
			this.transport = transport;
			this.transport.setRegistry(new FClientRegistry());
			this.protocolFactory = protocolFactory;
			this.inputProtocol = this.protocolFactory.getProtocol(this.transport);
			this.outputProtocol = this.protocolFactory.getProtocol(this.transport);
		}

		/**
		 * Use this to ping the server.
		 */
		public void ping(FContext ctx) throws TException {
			FProtocol oprot = this.outputProtocol;
			BlockingQueue<Object> result = new ArrayBlockingQueue<>(1);
			this.transport.register(ctx, recvPingHandler(ctx, result));
			try {
				synchronized (WRITE_LOCK) {
					oprot.writeRequestHeader(ctx);
					oprot.writeMessageBegin(new TMessage("ping", TMessageType.CALL, 0));
					Blah.ping_args args = new Blah.ping_args();
					args.write(oprot);
					oprot.writeMessageEnd();
					oprot.getTransport().flush();
				}

				Object res = null;
				try {
					res = result.poll(ctx.getTimeout(), TimeUnit.MILLISECONDS);
				} catch (InterruptedException e) {
					throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "ping interrupted: " + e.getMessage());
				}
				if (res == null) {
					throw new FTimeoutException("ping timed out");
				}
				if (res instanceof TException) {
					throw (TException) res;
				}
				Blah.ping_result r = (Blah.ping_result) res;
			} finally {
				this.transport.unregister(ctx);
			}
		}

		private FAsyncCallback recvPingHandler(final FContext ctx, final BlockingQueue<Object> result) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = Client.this.protocolFactory.getProtocol(tr);
					try {
						iprot.readResponseHeader(ctx);
						TMessage message = iprot.readMessageBegin();
						if (!message.name.equals("ping")) {
							throw new TApplicationException(TApplicationException.WRONG_METHOD_NAME, "ping failed: wrong method name");
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
									throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "ping interrupted: " + ie.getMessage());
								}
							}
							try {
								result.put(e);
							} finally {
								throw e;
							}
						}
						if (message.type != TMessageType.REPLY) {
							throw new TApplicationException(TApplicationException.INVALID_MESSAGE_TYPE, "ping failed: invalid message type");
						}
						Blah.ping_result res = new Blah.ping_result();
						res.read(iprot);
						iprot.readMessageEnd();
						try {
							result.put(res);
						} catch (InterruptedException e) {
							throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "ping interrupted: " + e.getMessage());
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

		/**
		 * Use this to tell the sever how you feel.
		 */
		public long bleh(FContext ctx, Thing one, Stuff Two) throws TException, InvalidOperation {
			FProtocol oprot = this.outputProtocol;
			BlockingQueue<Object> result = new ArrayBlockingQueue<>(1);
			this.transport.register(ctx, recvBlehHandler(ctx, result));
			try {
				synchronized (WRITE_LOCK) {
					oprot.writeRequestHeader(ctx);
					oprot.writeMessageBegin(new TMessage("bleh", TMessageType.CALL, 0));
					Blah.bleh_args args = new Blah.bleh_args();
					args.setOne(one);
					args.setTwo(Two);
					args.write(oprot);
					oprot.writeMessageEnd();
					oprot.getTransport().flush();
				}

				Object res = null;
				try {
					res = result.poll(ctx.getTimeout(), TimeUnit.MILLISECONDS);
				} catch (InterruptedException e) {
					throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "bleh interrupted: " + e.getMessage());
				}
				if (res == null) {
					throw new FTimeoutException("bleh timed out");
				}
				if (res instanceof TException) {
					throw (TException) res;
				}
				Blah.bleh_result r = (Blah.bleh_result) res;
				if (r.isSetSuccess()) {
					return r.success;
				}
				if (r.oops != null) {
					throw r.oops;
				}
				throw new TApplicationException(TApplicationException.MISSING_RESULT, "bleh failed: unknown result");
			} finally {
				this.transport.unregister(ctx);
			}
		}

		private FAsyncCallback recvBlehHandler(final FContext ctx, final BlockingQueue<Object> result) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = Client.this.protocolFactory.getProtocol(tr);
					try {
						iprot.readResponseHeader(ctx);
						TMessage message = iprot.readMessageBegin();
						if (!message.name.equals("bleh")) {
							throw new TApplicationException(TApplicationException.WRONG_METHOD_NAME, "bleh failed: wrong method name");
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
									throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "bleh interrupted: " + ie.getMessage());
								}
							}
							try {
								result.put(e);
							} finally {
								throw e;
							}
						}
						if (message.type != TMessageType.REPLY) {
							throw new TApplicationException(TApplicationException.INVALID_MESSAGE_TYPE, "bleh failed: invalid message type");
						}
						Blah.bleh_result res = new Blah.bleh_result();
						res.read(iprot);
						iprot.readMessageEnd();
						try {
							result.put(res);
						} catch (InterruptedException e) {
							throw new TApplicationException(TApplicationException.INTERNAL_ERROR, "bleh interrupted: " + e.getMessage());
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

		public Processor(Iface iface) {
			super(getProcessMap(iface, new java.util.HashMap<String, FProcessorFunction>()));
		}

		protected Processor(Iface iface, java.util.Map<String, FProcessorFunction> processMap) {
			super(getProcessMap(iface, processMap));
		}

		private static java.util.Map<String, FProcessorFunction> getProcessMap(Iface handler, java.util.Map<String, FProcessorFunction> processMap) {
			processMap.put("ping", new Ping(handler));
			processMap.put("bleh", new Bleh(handler));
			return processMap;
		}

		private static class Ping implements FProcessorFunction {

			private Iface handler;

			public Ping(Iface handler) {
				this.handler = handler;
			}

			public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {
				Blah.ping_args args = new Blah.ping_args();
				try {
					args.read(iprot);
				} catch (TException e) {
					iprot.readMessageEnd();
					synchronized (WRITE_LOCK) {
						writeApplicationException(ctx, oprot, TApplicationException.PROTOCOL_ERROR, "ping", e.getMessage());
					}
					throw e;
				}

				iprot.readMessageEnd();
				Blah.ping_result result = new Blah.ping_result();
				try {
					this.handler.ping(ctx);
				} catch (TException e) {
					synchronized (WRITE_LOCK) {
						writeApplicationException(ctx, oprot, TApplicationException.INTERNAL_ERROR, "ping", "Internal error processing ping: " + e.getMessage());
					}
					throw e;
				}
				synchronized (WRITE_LOCK) {
					try {
						oprot.writeResponseHeader(ctx);
						oprot.writeMessageBegin(new TMessage("ping", TMessageType.REPLY, 0));
						result.write(oprot);
						oprot.writeMessageEnd();
						oprot.getTransport().flush();
					} catch (TException e) {
						if (e instanceof FMessageSizeException) {
							writeApplicationException(ctx, oprot, FTransport.RESPONSE_TOO_LARGE, "ping", "response too large: " + e.getMessage());
						} else {
							throw e;
						}
					}
				}
			}
		}

		private static class Bleh implements FProcessorFunction {

			private Iface handler;

			public Bleh(Iface handler) {
				this.handler = handler;
			}

			public void process(FContext ctx, FProtocol iprot, FProtocol oprot) throws TException {
				Blah.bleh_args args = new Blah.bleh_args();
				try {
					args.read(iprot);
				} catch (TException e) {
					iprot.readMessageEnd();
					synchronized (WRITE_LOCK) {
						writeApplicationException(ctx, oprot, TApplicationException.PROTOCOL_ERROR, "bleh", e.getMessage());
					}
					throw e;
				}

				iprot.readMessageEnd();
				Blah.bleh_result result = new Blah.bleh_result();
				try {
					result.success = this.handler.bleh(ctx, args.one, args.Two);
					result.setSuccessIsSet(true);
				} catch (InvalidOperation oops) {
					result.oops = oops;
				} catch (TException e) {
					synchronized (WRITE_LOCK) {
						writeApplicationException(ctx, oprot, TApplicationException.INTERNAL_ERROR, "bleh", "Internal error processing bleh: " + e.getMessage());
					}
					throw e;
				}
				synchronized (WRITE_LOCK) {
					try {
						oprot.writeResponseHeader(ctx);
						oprot.writeMessageBegin(new TMessage("bleh", TMessageType.REPLY, 0));
						result.write(oprot);
						oprot.writeMessageEnd();
						oprot.getTransport().flush();
					} catch (TException e) {
						if (e instanceof FMessageSizeException) {
							writeApplicationException(ctx, oprot, FTransport.RESPONSE_TOO_LARGE, "bleh", "response too large: " + e.getMessage());
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
