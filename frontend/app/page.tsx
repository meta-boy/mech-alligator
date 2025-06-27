"use client";

import { AuthManager } from "@/utils/auth";
import { useEffect, useState } from "react";

interface Product {
  id: string;
  name: string;
  price: number;
  currency: string;
}

export default function Home() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const authManager = AuthManager.getInstance();
        const response = await authManager.makeAuthenticatedRequest(
          "/api/products"
        );

        if (!response.ok) {
          throw new Error("Failed to fetch products");
        }

        const data = await response.json();
        setProducts(data.products);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("An unknown error occurred");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  if (loading) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center p-24">
        <div className="text-center">
          <p className="text-lg leading-8">Loading products...</p>
        </div>
      </main>
    );
  }

  if (error) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center p-24">
        <div className="text-center">
          <h1 className="text-4xl font-bold tracking-tight text-red-500 sm:text-6xl">
            Error
          </h1>
          <p className="mt-6 text-lg leading-8">{error}</p>
        </div>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen flex-col items-center p-8">
      <h1 className="text-4xl font-bold tracking-tight sm:text-6xl mb-8">
        Our Products
      </h1>
      <div className="grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
        {products.map((product) => (
          <div
            key={product.id}
            className="rounded-lg border p-4 shadow-md"
          >
            <h2 className="text-xl font-bold">{product.name}</h2>
            <p className="mt-2 text-lg">
              {new Intl.NumberFormat("en-IN", {
                style: "currency",
                currency: product.currency,
              }).format(product.price / 100)}
            </p>
          </div>
        ))}
      </div>
    </main>
  );
}
