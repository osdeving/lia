package com.example.orders.application;

import com.example.orders.domain.OrderId;
import java.util.Objects;

public record CreateOrderInput(OrderId id) {
  public CreateOrderInput {
    Objects.requireNonNull(id, "id is required");
  }
}
