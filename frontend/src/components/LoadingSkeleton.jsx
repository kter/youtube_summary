function LoadingSkeleton() {
  return (
    <div className="glass-card rounded-2xl overflow-hidden">
      {/* Thumbnail skeleton */}
      <div className="skeleton w-full aspect-video"></div>
      
      {/* Content skeleton */}
      <div className="p-5 space-y-4">
        {/* Title */}
        <div className="space-y-2">
          <div className="skeleton h-5 w-full rounded"></div>
          <div className="skeleton h-5 w-3/4 rounded"></div>
        </div>
        
        {/* Channel */}
        <div className="skeleton h-4 w-1/3 rounded"></div>
        
        {/* Stats */}
        <div className="flex gap-4">
          <div className="skeleton h-4 w-16 rounded"></div>
          <div className="skeleton h-4 w-16 rounded"></div>
          <div className="skeleton h-4 w-20 rounded"></div>
        </div>
        
        {/* Summary */}
        <div className="pt-4 border-t border-slate-700/50 space-y-2">
          <div className="skeleton h-4 w-full rounded"></div>
          <div className="skeleton h-4 w-full rounded"></div>
          <div className="skeleton h-4 w-5/6 rounded"></div>
        </div>
      </div>
    </div>
  );
}

export default LoadingSkeleton;
