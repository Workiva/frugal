/*
 * Copyright 2017 Workiva
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

part of frugal.src.frugal;

/// Processes a method invocation on a proxied method and returns the result.
/// The arguments should match the arity of the proxied method, and have the
/// same types. The first argument will always be the [FContext].
typedef Future InvocationHandler(
    String serviceName, String methodName, List<Object> args);

/// Implements interceptor logic around API calls. This can be used, for
/// example, to implement retry policies on service calls, logging, telemetry,
/// or authentication and authorization. Middleware may be applied to either RPC
/// services or pub/sub scopes. Middleware returns an [InvocationHandler] which
/// proxies the given [InvocationHandler].
typedef InvocationHandler Middleware(InvocationHandler handler);

/// Contains an [InvocationHandler] used to proxy the given service method
/// This should only be used by generated code.
class FMethod {
  String _serviceName;
  String _methodName;
  InvocationHandler _handler;

  /// Creates an [FMethod] with the given function, service name, and method
  /// name.
  FMethod(dynamic f, String serviceName, String methodName,
      List<Middleware> middleware) {
    this._serviceName = serviceName;
    this._methodName = methodName;
    this._handler = _composeMiddleware(f, middleware);
  }

  /// Invokes the proxied [InvocationHandler] with the given arguments and
  /// returns the results.
  Future call(List<Object> args) {
    return this._handler(this._serviceName, this._methodName, args);
  }

  /// Applies the [Middleware] to the provided method.
  InvocationHandler _composeMiddleware(dynamic f, List<Middleware> middleware) {
    InvocationHandler handler =
        (String serviceName, String methodName, List<Object> args) {
      Future actual = Function.apply(f, args);
      return actual;
    };

    if (middleware == null) {
      return handler;
    }
    // ignore: STRONG_MODE_DOWN_CAST_COMPOSITE
    return middleware.fold(handler, (prev, element) => element(prev));
  }
}

/// [Middleware] for debugging that logs the requests and responses in json
/// format.
InvocationHandler debugMiddleware(InvocationHandler next) {
  return (String serviceName, String methodName, List<Object> args) async {
    // Logging the request in one block and the request + response in another
    // block so that it's easier to see what is happening. Indented for visual
    // clarity.
    List<String> requestLog = [];
    List<String> responseLog = [];
    requestLog.add('frugal request to $serviceName.$methodName');
    responseLog.add('frugal response from $serviceName.$methodName');
    for (int i = 0; i < args.length; i++) {
      int iHuman = i + 1;
      var obj = args[i];
      String type = obj.runtimeType.toString();
      String json = fObjToJson(obj);
      String argString = '  arg #$iHuman: $type: $json';
      requestLog.add(argString);
      responseLog.add(argString);
    }
    print(requestLog.join('\n'));
    Object ret = await next(serviceName, methodName, args);
    String type = ret.runtimeType.toString();
    String json = fObjToJson(ret);
    responseLog.add('response: $type: $json');
    print(responseLog.join('\n'));
    return ret;
  };
}
