package com.example.orders.port;

import com.example.orders.domain.OrderId;

public interface OrderRepository {
  boolean save(OrderId id);
}
