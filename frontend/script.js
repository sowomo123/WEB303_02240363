const productInput = document.getElementById("productId");
const message = document.getElementById("message");
const productResult = document.getElementById("productResult");
const productImage = document.getElementById("productImage");
const searchButton = document.getElementById("searchButton");
const productImageClasses = {
    P001: "laptop-image",
    P002: "keyboard-image",
    P003: "mouse-image"
};

async function getProduct(productId = productInput.value.trim()) {
    message.textContent = "";
    message.className = "";
    productResult.classList.add("hidden");

    if (productId === "") {
        message.textContent = "Please enter a product ID";
        productInput.focus();
        return;
    }

    searchButton.disabled = true;
    searchButton.innerHTML = "Searching...";

    try {
        const response = await fetch(
            `/product/${encodeURIComponent(productId)}`
        );

        const data = await response.json();

        if (!response.ok) {
            if (response.status === 503 || data.error === "Product service is unavailable.") {
                message.textContent = "Product service is unavailable.\nPlease try again later.";
                return;
            }
            message.textContent = data.error || "An error occurred.";
            return;
        }

        document.getElementById("resultId").textContent = data.product_id;
        document.getElementById("resultName").textContent = data.name;
        document.getElementById("resultPrice").textContent = formatPrice(data.price);
        productImage.className = "product-image";
        const imageClass = productImageClasses[data.product_id.toUpperCase()];
        if (imageClass) productImage.classList.add(imageClass);
        productResult.classList.remove("hidden");
        message.textContent = "Product found in the catalog.";
        message.className = "success";
    } catch (error) {
        message.textContent = "Product service is unavailable.\nPlease try again later.";
    } finally {
        searchButton.disabled = false;
        searchButton.innerHTML = 'Search <span aria-hidden="true">↗</span>';
    }
}

function formatPrice(price) {
    const amount = Number(price);
    return Number.isFinite(amount) ? `$${amount.toFixed(2)}` : price;
}

function showAddedMessage() {
    message.textContent = "Added to your bag.";
    message.className = "success";
}

productInput.addEventListener("keydown", (event) => {
    if (event.key === "Enter") getProduct();
});

document.querySelectorAll("[data-product-id]").forEach((button) => {
    button.addEventListener("click", () => {
        productInput.value = button.dataset.productId;
        getProduct(button.dataset.productId);
    });
});