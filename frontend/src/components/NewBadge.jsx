import { Sparkles } from 'lucide-react';

function NewBadge({ processedAt }) {
  // Check if processed within last 24 hours
  const isNew = () => {
    if (!processedAt) return false;
    
    const processedDate = new Date(processedAt);
    const now = new Date();
    const hoursDiff = (now - processedDate) / (1000 * 60 * 60);
    
    return hoursDiff <= 24;
  };

  if (!isNew()) return null;

  return (
    <div className="absolute top-3 left-3">
      <span className="badge-new inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-semibold bg-gradient-to-r from-accent-500 to-primary-500 text-white shadow-lg shadow-accent-500/30">
        <Sparkles className="h-3 w-3" />
        NEW
      </span>
    </div>
  );
}

export default NewBadge;
