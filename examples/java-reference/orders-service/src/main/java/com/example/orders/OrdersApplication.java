package com.example.orders;

import com.example.orders.application.CreateOrderInput;
import com.example.orders.bootstrap.OrdersWiring;
import com.example.orders.domain.OrderId;

public final class OrdersApplication {
  private OrdersApplication() {
  }

  public static void main(String[] args) {
    var wiring = new OrdersWiring();
    var useCase = wiring.createOrderUseCase();
    var output = useCase.execute(new CreateOrderInput(new OrderId("ord-001")));
    System.out.println("Created order: " + output.accepted());
  }
}
