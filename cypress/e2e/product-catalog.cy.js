describe("Product catalog", () => {
  const product = {
    product_id: "P001",
    name: "Laptop",
    price: 1299.5
  };

  beforeEach(() => {
    cy.visit("/");
  });

  it("loads the catalog page", () => {
    cy.contains("Find something").should("be.visible");
    cy.get("#productId").should("be.visible");
    cy.get("#searchButton").should("contain", "Search");
  });

  it("shows a validation message for an empty product ID", () => {
    cy.get("#searchButton").click();
    cy.get("#message").should("have.text", "Please enter a product ID");
    cy.get("#productResult").should("have.class", "hidden");
    cy.get("#productId").should("be.focused");
  });

  it("finds a product by ID", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 200,
      body: product
    }).as("getProduct");

    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@getProduct");

    cy.get("#message").should("have.text", "Product found in the catalog.");
    cy.get("#productResult").should("not.have.class", "hidden");
    cy.get("#resultId").should("have.text", "P001");
    cy.get("#resultName").should("have.text", "Laptop");
    cy.get("#resultPrice").should("have.text", "$1299.50");
  });

  it("searches when Enter is pressed in the product field", () => {
    cy.intercept("GET", "/product/P002", {
      statusCode: 200,
      body: { product_id: "P002", name: "Keyboard", price: 89 }
    }).as("getProduct");

    cy.get("#productId").type("P002{enter}");
    cy.wait("@getProduct").its("request.url").should("include", "/product/P002");
    cy.get("#resultName").should("have.text", "Keyboard");
  });

  it("searches a product from a popular-product shortcut", () => {
    cy.intercept("GET", "/product/P003", {
      statusCode: 200,
      body: { product_id: "P003", name: "Mouse", price: 24.99 }
    }).as("getProduct");

    cy.contains("button", "P003").click();
    cy.wait("@getProduct");

    cy.get("#productId").should("have.value", "P003");
    cy.get("#resultId").should("have.text", "P003");
    cy.get("#resultName").should("have.text", "Mouse");
  });

  it("shows a not-found message for an unknown product", () => {
    cy.intercept("GET", "/product/UNKNOWN", {
      statusCode: 404,
      body: { error: "Product not found." }
    }).as("getProduct");

    cy.get("#productId").type("UNKNOWN");
    cy.get("#searchButton").click();
    cy.wait("@getProduct");

    cy.get("#message").should("have.text", "Product not found.");
    cy.get("#productResult").should("have.class", "hidden");
  });

  it("shows the unavailable message for a 503 response", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 503,
      body: { error: "Product service is unavailable." }
    }).as("getProduct");

    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@getProduct");

    cy.get("#message").should("contain", "Product service is unavailable.");
    cy.get("#productResult").should("have.class", "hidden");
  });

  it("shows the unavailable message when the request fails", () => {
    cy.intercept("GET", "/product/P001", { forceNetworkError: true }).as("getProduct");

    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@getProduct");

    cy.get("#message").should("contain", "Product service is unavailable.");
  });

  it("disables the search button while the request is pending", () => {
    cy.intercept("GET", "/product/P001", (request) => {
      request.reply({
        delay: 100,
        statusCode: 200,
        body: product
      });
    }).as("getProduct");

    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.get("#searchButton").should("be.disabled").and("contain", "Searching");
    cy.wait("@getProduct");
    cy.get("#searchButton").should("not.be.disabled").and("contain", "Search");
  });

  it("shows confirmation after adding a product to the bag", () => {
    cy.intercept("GET", "/product/P001", {
      statusCode: 200,
      body: product
    }).as("getProduct");

    cy.get("#productId").type("P001");
    cy.get("#searchButton").click();
    cy.wait("@getProduct");
    cy.get(".add-button").click();

    cy.get("#message").should("have.text", "Added to your bag.");
    cy.get("#message").should("have.class", "success");
  });
});
