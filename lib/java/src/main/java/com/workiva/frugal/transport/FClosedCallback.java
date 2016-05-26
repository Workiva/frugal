package com.workiva.frugal.transport;

/**
 * When a {@code FTransport} is closed for any reason, the {@code FTransport}
 * object's {@code FClosedCallback} is notified, if one has been registered.
 *
 * @deprecated use {@code FTransportClosedCallback} instead.
 */
@Deprecated
public interface FClosedCallback {

    /**
     * This callback notification method is invoked when the {@code FTransport} is
     * closed.
     */
    void onClose();

}

