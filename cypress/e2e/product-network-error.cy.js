describe("Product network failure", () => {
  it("shows an error when the request fails", () => {
    cy.intercept("GET", "/product/P001", {
      forceNetworkError: true
    }).as("product");

    cy.visit("/");
    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@product");

    cy.get("#message").should(
      "contain",
      "Product service is unavailable."
    );
  });
});
