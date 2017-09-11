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

/// Watches and heals an [FTransport]. Exposes a number of hooks which can be
/// used to add logic around [FTransport] events, such as unexpected
/// disconnects, expected disconnects, failed reconnects, and successful
/// reconnects.
///
/// [BaseFTransportMonitor] is a basic implementation with backoffs and max
/// attempts. This can be extended or reimplemented to provide custom logic.
abstract class FTransportMonitor {
  /// Called when the transport is closed cleanly by a call to [FTransport]
  /// close with no error.
  void onClosedCleanly();

  /// Called when the transport is closed for a reason other than a clean call
  /// to [FTransport] close. Returns the number of milliseconds to wait before
  /// attempting to reopen the transport or a negative number indicating not to
  /// reopen.
  int onClosedUncleanly(Object cause);

  /// Called when an attempt to reopen the transport fails. Returns the number
  /// of milliseconds to wait before attempting to reopen the transport. A
  /// negative value means the transport will not attempt to be reopened.
  int onReopenFailed(int prevAttempts, int prevWait);

  /// Called after the transport has been successfully reopened.
  void onReopenSucceeded();
}
