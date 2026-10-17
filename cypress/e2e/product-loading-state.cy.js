describe("Product search loading state", () => {
  it("disables the search button while loading", () => {
    cy.intercept("GET", "/product/P001", {
      delay: 300,
      statusCode: 200,
      body: { product_id: "P001", name: "Laptop", price: 1299.5 }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();

    cy.get("#searchButton")
      .should("be.disabled")
      .and("contain", "Searching");

    cy.wait("@product");

    cy.get("#searchButton")
      .should("not.be.disabled")
      .and("contain", "Search");
  });
});
