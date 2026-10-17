describe("Product result details", () => {
  it("displays the returned product details", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 200,
      body: { product_id: "P001", name: "Laptop", price: 1299.5 }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@product");

    cy.get("#resultId").should("have.text", "P001");
    cy.get("#resultName").should("have.text", "Laptop");
    cy.get("#resultPrice").should("have.text", "$1299.50");
  });
});
