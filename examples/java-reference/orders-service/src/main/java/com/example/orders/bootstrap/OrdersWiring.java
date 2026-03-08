package com.example.orders.bootstrap;

import com.example.orders.application.CreateOrderUseCase;
import com.example.orders.infrastructure.InMemoryOrderRepository;
import com.example.orders.port.OrderRepository;

public final class OrdersWiring {
  private final OrderRepository orderRepository;

  public OrdersWiring() {
    this.orderRepository = new InMemoryOrderRepository();
  }

  public OrderRepository orderRepository() {
    return orderRepository;
  }

  public CreateOrderUseCase createOrderUseCase() {
    return new CreateOrderUseCase(orderRepository);
  }
}
