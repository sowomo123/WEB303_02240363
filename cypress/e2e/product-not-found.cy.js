describe("Unknown product", () => {
  it("displays a not-found message", () => {
    cy.intercept("GET", "/product/UNKNOWN", {
      statusCode: 404,
      body: { error: "Product not found." }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("UNKNOWN");
    cy.get("#searchButton").click();
    cy.wait("@product");

    cy.get("#message").should("have.text", "Product not found.");
    cy.get("#productResult").should("have.class", "hidden");
  });
});
