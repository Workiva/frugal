/**
 * Autogenerated by Frugal Compiler (2.0.0-RC5)
 * DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
 *
 * @generated
 */

package v1.music;

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

import java.util.Arrays;
import java.util.List;
import java.util.logging.Logger;
import javax.annotation.Generated;




@Generated(value = "Autogenerated by Frugal Compiler (2.0.0-RC5)")
public class AlbumWinnersSubscriber {

	/**
	 * Scopes are a Frugal extension to the IDL for declaring PubSub
	 * semantics. Subscribers to this scope will be notified if they win a contest.
	 * Scopes must have a prefix.
	 */
	public interface Iface {
		public FSubscription subscribeWinner(final WinnerHandler handler) throws TException;

	}

	public interface WinnerHandler {
		void onWinner(FContext ctx, Album req);
	}

	/**
	 * Scopes are a Frugal extension to the IDL for declaring PubSub
	 * semantics. Subscribers to this scope will be notified if they win a contest.
	 * Scopes must have a prefix.
	 */
	public static class Client implements Iface {
		private static final String DELIMITER = ".";
		private static final Logger LOGGER = Logger.getLogger(Client.class.getName());

		private final FScopeProvider provider;
		private final ServiceMiddleware[] middleware;

		public Client(FScopeProvider provider, ServiceMiddleware... middleware) {
			this.provider = provider;
			List<ServiceMiddleware> combined = middleware;
			combined.addAll(Arrays.asList(provider.getMiddleware()));
			this.middleware = combined.toArray(new ServiceMiddleware[0]);
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
						throw new TApplicationException(TApplicationException.UNKNOWN_METHOD);
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
