package com.workiva.frugal.transport;

import org.apache.thrift.TConfiguration;
import static org.junit.Assert.assertEquals;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.junit.runners.JUnit4;

@RunWith(JUnit4.class)
public class TConfigurationBuilderTest {

  @Test
  public void setMaxMessageSize_withPositiveValue() {
    TConfiguration config = TConfigurationBuilder.custom().setMaxMessageSize(32).build();
    assertEquals(32, config.getMaxMessageSize());
  }

  @Test
  public void setMaxMessageSize_withNegativeValue() {
    TConfiguration config = TConfigurationBuilder.custom().setMaxMessageSize(-27).build();
    assertEquals(Integer.MAX_VALUE, config.getMaxMessageSize());
  }

  @Test
  public void setMaxMessageSize_withZeroValue() {
    TConfiguration config = TConfigurationBuilder.custom().setMaxMessageSize(0).build();
    assertEquals(Integer.MAX_VALUE, config.getMaxMessageSize());
  }
}
