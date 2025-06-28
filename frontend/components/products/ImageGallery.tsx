"use client";

import { useEffect, useState } from 'react';
import Image from 'next/image';
import Head from 'next/head';
import { ChevronLeft, ChevronRight, Package } from 'lucide-react';

interface ImageGalleryProps {
  images: string[];
  productName: string;
}

export const ImageGallery = ({ images, productName }: ImageGalleryProps) => {
  const getRandomImages = (imageArray: string[], name: string, count: number = 3) => {
    if (!imageArray || imageArray.length === 0) return [];
    
    let hash = 0;
    for (let i = 0; i < name.length; i++) {
      const char = name.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
    
    const selectedImages = [];
    const availableImages = [...imageArray];
    const selectCount = Math.min(count, availableImages.length);
    
    for (let i = 0; i < selectCount; i++) {
      const index = Math.abs(hash + i) % availableImages.length;
      selectedImages.push(availableImages[index]);
      availableImages.splice(index, 1);
    }
    
    return selectedImages;
  };

  const selectedImages = getRandomImages(images, productName, 3);
  const [currentImageIndex, setCurrentImageIndex] = useState(0);

  useEffect(() => {
    selectedImages.forEach((src) => {
      const img = new window.Image();
      img.src = src;
    });
  }, [selectedImages]);
  
  const nextImage = (e: React.MouseEvent) => {
    e.stopPropagation();
    setCurrentImageIndex((prev) => (prev + 1) % selectedImages.length);
  };
  
  const prevImage = (e: React.MouseEvent) => {
    e.stopPropagation();
    setCurrentImageIndex((prev) => (prev - 1 + selectedImages.length) % selectedImages.length);
  };

  if (!selectedImages || selectedImages.length === 0) {
    return (
      <div className="h-48 bg-gradient-to-br from-slate-100 to-slate-200 relative flex items-center justify-center">
        <Package className="h-16 w-16 text-slate-400" />
      </div>
    );
  }

  return (
    <>
      <Head>
        {selectedImages.map((src, index) => (
          <link key={`preload-${index}`} rel="preload" href={src} as="image" />
        ))}
      </Head>
      <div className="h-48 bg-gradient-to-br from-slate-100 to-slate-200 relative overflow-hidden group">
        <div className="w-full h-full">
          <Image
            src={selectedImages[currentImageIndex]}
            alt={`${productName} - Image ${currentImageIndex + 1}`}
            fill
            className="object-cover transition-all duration-300"
            style={{ objectFit: "cover" }}
            quality={60}
            onError={(e) => {
              const target = e.target as HTMLImageElement;
              target.style.display = 'none';
              target.nextElementSibling?.classList.remove('hidden');
            }}
            sizes="(max-width: 768px) 100vw, 400px"
            priority={currentImageIndex === 0}
          />
        </div>
        
        <div className="absolute inset-0 hidden bg-gradient-to-br from-slate-100 to-slate-200 flex items-center justify-center">
          <Package className="h-16 w-16 text-slate-400" />
        </div>

        {selectedImages.length > 1 && (
          <>
            <button
              onClick={prevImage}
              className="absolute left-2 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white p-1 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button
              onClick={nextImage}
              className="absolute right-2 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white p-1 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
            
            <div className="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-1">
              {selectedImages.map((_, index) => (
                <button
                  key={index}
                  onClick={(e) => {
                    e.stopPropagation();
                    setCurrentImageIndex(index);
                  }}
                  className={`w-2 h-2 rounded-full transition-all duration-200 ${
                    index === currentImageIndex 
                      ? 'bg-white' 
                      : 'bg-white/50 hover:bg-white/70'
                  }`}
                />
              ))}
            </div>
          </>
        )}
      </div>
    </>
  );
};