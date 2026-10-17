describe("Product search using Enter", () => {
  it("searches when Enter is pressed", () => {
    cy.intercept("GET", "/product/P002", {
      statusCode: 200,
      body: { product_id: "P002", name: "Keyboard", price: 89 }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P002{enter}");
    cy.wait("@product");

    cy.get("#resultName").should("have.text", "Keyboard");
  });
});
