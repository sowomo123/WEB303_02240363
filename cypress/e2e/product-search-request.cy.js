describe("Product search request", () => {
  it("sends a GET request for the entered product ID", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 200,
      body: { product_id: "P001", name: "Laptop", price: 1299.5 }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();

    cy.wait("@product").its("request.method").should("equal", "GET");
    cy.get("#resultId").should("have.text", "P001");
  });
});
