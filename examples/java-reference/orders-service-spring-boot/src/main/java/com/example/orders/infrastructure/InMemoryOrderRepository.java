package com.example.orders.infrastructure;

import com.example.orders.domain.OrderId;
import com.example.orders.port.OrderRepository;
import java.util.HashSet;
import java.util.Set;

public final class InMemoryOrderRepository implements OrderRepository {
  private final Set<OrderId> orders = new HashSet<>();

  @Override
  public boolean save(OrderId id) {
    return orders.add(id);
  }
}
