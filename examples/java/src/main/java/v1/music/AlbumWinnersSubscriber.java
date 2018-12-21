/**
 * Autogenerated by Frugal Compiler (2.25.2)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *
 * @generated
 */

package v1.music;

import com.workiva.frugal.FContext;
import com.workiva.frugal.exception.TApplicationExceptionType;
import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import com.workiva.frugal.protocol.*;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.transport.FPublisherTransport;
import com.workiva.frugal.transport.FSubscriberTransport;
import com.workiva.frugal.transport.FSubscription;
import com.workiva.frugal.transport.TMemoryOutputBuffer;
import org.apache.thrift.TException;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.transport.TTransport;
import org.apache.thrift.transport.TTransportException;
import org.apache.thrift.protocol.*;

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
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import javax.annotation.Generated;




@Generated(value = "Autogenerated by Frugal Compiler (2.25.2)")
public class AlbumWinnersSubscriber {

	/**
	 * Scopes are a Frugal extension to the IDL for declaring PubSub
	 * semantics. Subscribers to this scope will be notified if they win a contest.
	 * Scopes must have a prefix.
	 */
	public interface Iface {
		public FSubscription subscribeContestStart(final ContestStartHandler handler) throws TException;

		public FSubscription subscribeTimeLeft(final TimeLeftHandler handler) throws TException;

		public FSubscription subscribeWinner(final WinnerHandler handler) throws TException;

	}

	public interface IfaceThrowable {
		public FSubscription subscribeContestStartThrowable(final ContestStartThrowableHandler handler) throws TException;

		public FSubscription subscribeTimeLeftThrowable(final TimeLeftThrowableHandler handler) throws TException;

		public FSubscription subscribeWinnerThrowable(final WinnerThrowableHandler handler) throws TException;

	}

	public interface ContestStartHandler {
		void onContestStart(FContext ctx, java.util.List<Album> req) throws TException;
	}

	public interface TimeLeftHandler {
		void onTimeLeft(FContext ctx, double req) throws TException;
	}

	public interface WinnerHandler {
		void onWinner(FContext ctx, Album req) throws TException;
	}

	public interface ContestStartThrowableHandler {
		void onContestStart(FContext ctx, java.util.List<Album> req) throws TException;
	}

	public interface TimeLeftThrowableHandler {
		void onTimeLeft(FContext ctx, double req) throws TException;
	}

	public interface WinnerThrowableHandler {
		void onWinner(FContext ctx, Album req) throws TException;
	}

	/**
	 * Scopes are a Frugal extension to the IDL for declaring PubSub
	 * semantics. Subscribers to this scope will be notified if they win a contest.
	 * Scopes must have a prefix.
	 */
	public static class Client implements Iface, IfaceThrowable {
		private static final String DELIMITER = ".";
		private static final Logger LOGGER = LoggerFactory.getLogger(Client.class);

		private final FScopeProvider provider;
		private final ServiceMiddleware[] middleware;

		public Client(FScopeProvider provider, ServiceMiddleware... middleware) {
			this.provider = provider;
			List<ServiceMiddleware> combined = Arrays.asList(middleware);
			combined.addAll(provider.getMiddleware());
			this.middleware = combined.toArray(new ServiceMiddleware[0]);
		}

		public FSubscription subscribeContestStart(final ContestStartHandler handler) throws TException {
			final String op = "ContestStart";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final ContestStartHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, ContestStartHandler.class, middleware);
			transport.subscribe(topic, recvContestStart(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvContestStart(String op, FProtocolFactory pf, ContestStartHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					org.apache.thrift.protocol.TList elem42 = iprot.readListBegin();
					java.util.List<Album> received = new ArrayList<Album>(elem42.size);
					for (int elem43 = 0; elem43 < elem42.size; ++elem43) {
						Album elem44 = new Album();
						elem44.read(iprot);
						received.add(elem44);
					}
					iprot.readListEnd();
					iprot.readMessageEnd();
					handler.onContestStart(ctx, received);
				}
			};
		}

		public FSubscription subscribeTimeLeft(final TimeLeftHandler handler) throws TException {
			final String op = "TimeLeft";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final TimeLeftHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, TimeLeftHandler.class, middleware);
			transport.subscribe(topic, recvTimeLeft(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvTimeLeft(String op, FProtocolFactory pf, TimeLeftHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					double received = iprot.readDouble();
					iprot.readMessageEnd();
					handler.onTimeLeft(ctx, received);
				}
			};
		}

		public FSubscription subscribeWinner(final WinnerHandler handler) throws TException {
			final String op = "Winner";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final WinnerHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, WinnerHandler.class, middleware);
			transport.subscribe(topic, recvWinner(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvWinner(String op, FProtocolFactory pf, WinnerHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					Album received = new Album();
					received.read(iprot);
					iprot.readMessageEnd();
					handler.onWinner(ctx, received);
				}
			};
		}

		public FSubscription subscribeContestStartThrowable(final ContestStartThrowableHandler handler) throws TException {
			final String op = "ContestStart";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final ContestStartThrowableHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, ContestStartThrowableHandler.class, middleware);
			transport.subscribe(topic, recvContestStart(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvContestStart(String op, FProtocolFactory pf, ContestStartThrowableHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					org.apache.thrift.protocol.TList elem45 = iprot.readListBegin();
					java.util.List<Album> received = new ArrayList<Album>(elem45.size);
					for (int elem46 = 0; elem46 < elem45.size; ++elem46) {
						Album elem47 = new Album();
						elem47.read(iprot);
						received.add(elem47);
					}
					iprot.readListEnd();
					iprot.readMessageEnd();
					handler.onContestStart(ctx, received);
				}
			};
		}

		public FSubscription subscribeTimeLeftThrowable(final TimeLeftThrowableHandler handler) throws TException {
			final String op = "TimeLeft";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final TimeLeftThrowableHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, TimeLeftThrowableHandler.class, middleware);
			transport.subscribe(topic, recvTimeLeft(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvTimeLeft(String op, FProtocolFactory pf, TimeLeftThrowableHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					double received = iprot.readDouble();
					iprot.readMessageEnd();
					handler.onTimeLeft(ctx, received);
				}
			};
		}

		public FSubscription subscribeWinnerThrowable(final WinnerThrowableHandler handler) throws TException {
			final String op = "Winner";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FSubscriberTransport transport = subscriber.getTransport();
			final WinnerThrowableHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, WinnerThrowableHandler.class, middleware);
			transport.subscribe(topic, recvWinner(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FAsyncCallback recvWinner(String op, FProtocolFactory pf, WinnerThrowableHandler handler) {
			return new FAsyncCallback() {
				public void onMessage(TTransport tr) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					Album received = new Album();
					received.read(iprot);
					iprot.readMessageEnd();
					handler.onWinner(ctx, received);
				}
			};
		}
	}

}
