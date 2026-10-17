describe("Product service unavailable", () => {
  it("shows an error when the service returns 503", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 503,
      body: { error: "Product service is unavailable." }
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@product");

    cy.get("#message").should(
      "contain",
      "Product service is unavailable."
    );
    cy.get("#productResult").should("have.class", "hidden");
  });
});
