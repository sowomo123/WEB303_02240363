describe("Add product to bag", () => {
  it("shows confirmation after adding a product", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 200,
      body: { product_id: "P001", name: "Laptop", price: 1299.5 }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@product");

    cy.get(".add-button").click();

    cy.get("#message").should("have.text", "Added to your bag.");
    cy.get("#message").should("have.class", "success");
  });
});
