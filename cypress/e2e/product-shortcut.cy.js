describe("Popular product shortcut", () => {
  it("searches for P003 when its shortcut is clicked", () => {
    cy.intercept("GET", "/product/P003", {
      statusCode: 200,
      body: { product_id: "P003", name: "Mouse", price: 24.99 }
    }).as("product");

    cy.visit("/");
    cy.contains("button", "P003").click();
    cy.wait("@product");

    cy.get("#productId").should("have.value", "P003");
    cy.get("#resultName").should("have.text", "Mouse");
  });
});
