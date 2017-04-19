/**
 * Autogenerated by Frugal Compiler (2.3.0-RC2)
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
import com.workiva.frugal.provider.FDurableScopeProvider;
import com.workiva.frugal.provider.FScopeProvider;
import com.workiva.frugal.transport.FDurablePublisherTransport;
import com.workiva.frugal.transport.FPublisherTransport;
import com.workiva.frugal.transport.FDurableSubscriberTransport;
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




@Generated(value = "Autogenerated by Frugal Compiler (2.3.0-RC2)")
public class AlbumWinnersDurableSubscriber {

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

	public interface ContestStartHandler {
		void onContestStart(FContext ctx, String groupId, java.util.List<Album> req);
	}

	public interface TimeLeftHandler {
		void onTimeLeft(FContext ctx, String groupId, double req);
	}

	public interface WinnerHandler {
		void onWinner(FContext ctx, String groupId, Album req);
	}

	/**
	 * Scopes are a Frugal extension to the IDL for declaring PubSub
	 * semantics. Subscribers to this scope will be notified if they win a contest.
	 * Scopes must have a prefix.
	 */
	public static class Client implements Iface {
		private static final String DELIMITER = ".";
		private static final Logger LOGGER = LoggerFactory.getLogger(Client.class);

		private final FDurableScopeProvider provider;
		private final ServiceMiddleware[] middleware;

		public Client(FDurableScopeProvider provider, ServiceMiddleware... middleware) {
			this.provider = provider;
			List<ServiceMiddleware> combined = Arrays.asList(middleware);
			combined.addAll(provider.getMiddleware());
			this.middleware = combined.toArray(new ServiceMiddleware[0]);
		}

		public FSubscription subscribeContestStart(final ContestStartHandler handler) throws TException {
			final String op = "ContestStart";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FDurableScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FDurableSubscriberTransport transport = subscriber.getTransport();
			final ContestStartHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, ContestStartHandler.class, middleware);
			transport.subscribe(topic, recvContestStart(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FDurableAsyncCallback recvContestStart(String op, FProtocolFactory pf, ContestStartHandler handler) {
			return new FDurableAsyncCallback() {
				public void onMessage(TTransport tr, String groupId) throws TException {
					FProtocol iprot = pf.getProtocol(tr);
					FContext ctx = iprot.readRequestHeader();
					TMessage msg = iprot.readMessageBegin();
					if (!msg.name.equals(op)) {
						TProtocolUtil.skip(iprot, TType.STRUCT);
						iprot.readMessageEnd();
						throw new TApplicationException(TApplicationExceptionType.UNKNOWN_METHOD);
					}
					org.apache.thrift.protocol.TList elem47 = iprot.readListBegin();
					java.util.List<Album> received = new ArrayList<Album>(elem47.size);
					for (int elem48 = 0; elem48 < elem47.size; ++elem48) {
						Album elem49 = new Album();
						elem49.read(iprot);
						received.add(elem49);
					}
					iprot.readListEnd();
					iprot.readMessageEnd();
					handler.onContestStart(ctx, groupId, received);
				}
			};
		}



		public FSubscription subscribeTimeLeft(final TimeLeftHandler handler) throws TException {
			final String op = "TimeLeft";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FDurableScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FDurableSubscriberTransport transport = subscriber.getTransport();
			final TimeLeftHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, TimeLeftHandler.class, middleware);
			transport.subscribe(topic, recvTimeLeft(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FDurableAsyncCallback recvTimeLeft(String op, FProtocolFactory pf, TimeLeftHandler handler) {
			return new FDurableAsyncCallback() {
				public void onMessage(TTransport tr, String groupId) throws TException {
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
					handler.onTimeLeft(ctx, groupId, received);
				}
			};
		}



		public FSubscription subscribeWinner(final WinnerHandler handler) throws TException {
			final String op = "Winner";
			String prefix = "v1.music.";
			final String topic = String.format("%sAlbumWinners%s%s", prefix, DELIMITER, op);
			final FDurableScopeProvider.Subscriber subscriber = provider.buildSubscriber();
			final FDurableSubscriberTransport transport = subscriber.getTransport();
			final WinnerHandler proxiedHandler = InvocationHandler.composeMiddleware(handler, WinnerHandler.class, middleware);
			transport.subscribe(topic, recvWinner(op, subscriber.getProtocolFactory(), proxiedHandler));
			return FSubscription.of(topic, transport);
		}

		private FDurableAsyncCallback recvWinner(String op, FProtocolFactory pf, WinnerHandler handler) {
			return new FDurableAsyncCallback() {
				public void onMessage(TTransport tr, String groupId) throws TException {
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
					handler.onWinner(ctx, groupId, received);
				}
			};
		}

	}

}
