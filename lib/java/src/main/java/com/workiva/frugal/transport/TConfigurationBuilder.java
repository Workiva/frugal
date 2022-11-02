package com.workiva.frugal.transport;

import org.apache.thrift.TConfiguration;

public class TConfigurationBuilder {
  public static final TConfiguration DEFAULT = TConfiguration.DEFAULT;
  private final TConfiguration.Builder builder = TConfiguration.custom();

  public static TConfigurationBuilder custom() {
    return new TConfigurationBuilder();
  }

  public TConfigurationBuilder setMaxMessageSize(int maxMessageSize) {
    builder.setMaxMessageSize(maxMessageSize > 0 ? maxMessageSize : Integer.MAX_VALUE);
    return this;
  }

  public TConfiguration build() {
    return builder.build();
  }
}
