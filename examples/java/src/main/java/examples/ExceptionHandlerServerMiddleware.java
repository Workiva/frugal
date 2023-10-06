package examples;

import com.workiva.frugal.middleware.InvocationHandler;
import com.workiva.frugal.middleware.ServiceMiddleware;
import java.lang.reflect.InvocationTargetException;
import java.lang.reflect.Method;
import org.apache.thrift.TApplicationException;
import org.apache.thrift.TException;
import org.apache.thrift.transport.TTransportException;
import v1.music.PurchasingError;

public class ExceptionHandlerServerMiddleware implements ServiceMiddleware {

    @Override
    public <T> InvocationHandler<T> apply(T next) {
        return new InvocationHandler<T>(next) {
            @Override
            public Object invoke(Method method, Object receiver, Object[] args) throws Throwable {
                System.out.printf("==== CALLING %s.%s ====\n", method.getDeclaringClass().getName(), method.getName());

                // Default error code is UNKNOWN
                int errorCode = TApplicationException.UNKNOWN;
                try {
                    try {
                        Object result = method.invoke(receiver, args);
                        return result;
                    } catch (InvocationTargetException e) {
                        // Throw underlying exception
                        throw e.getTargetException();
                    }
                } catch (TApplicationException e) {
                    // See https://github.com/Workiva/frugal/blob/master/lib/java/src/main/java/com/workiva/frugal/exception/TApplicationExceptionType.java
                    // and https://github.com/Workiva/messaging-sdk/blob/master/lib/java/sdk/src/main/java/com/workiva/messaging_sdk/exception/TApplicationExceptionType.java
                    errorCode = e.getType();
                } catch (TTransportException e) {
                    // See https://github.com/Workiva/frugal/blob/master/lib/java/src/main/java/com/workiva/frugal/exception/TTransportExceptionType.java
                    // and https://github.com/Workiva/messaging-sdk/blob/master/lib/java/sdk/src/main/java/com/workiva/messaging_sdk/exception/TTransportExceptionType.java
                    errorCode = e.getType();
                } catch (PurchasingError e) {
                    // We can't really do this in a generic way because this is a user defined struct
                    // Best we can do is a generic error code
                    //
                    // Maybe that is okay for MVP and if folks want better granularity they can update their code or use this middleware
                    // as a base for their own (which isn't that uncommon)
                    errorCode = TApplicationException.UNKNOWN;
                }
                System.out.printf("==== CALLED  %s.%s ====\n", method.getDeclaringClass().getName(), method.getName());

                return null;
            }
        };
    }
}