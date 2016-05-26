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

