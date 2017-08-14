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

package com.workiva.frugal.transport;

/**
 * When a {@code FTransport} is closed for any reason, the {@code FTransport}
 * object's {@code FTransportClosedCallback} is notified, if one has been registered.
 */
public interface FTransportClosedCallback {

    /**
     * This callback notification method is invoked when the {@code FTransport} is
     * closed.
     *
     * @param cause the cause of the close or null if it was clean (resulting from a call to close()).
     */
    void onClose(Exception cause);

}

