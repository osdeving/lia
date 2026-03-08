package com.example.orders.domain;

import java.util.Objects;

public record OrderId(String value) {
  public OrderId {
    Objects.requireNonNull(value, "value is required");
    if (value.isBlank()) {
      throw new IllegalArgumentException("value must not be blank");
    }
  }
}
